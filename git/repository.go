package git

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/keys-pub/keys"

	"github.com/keys-pub/keys/keyring"
	git "github.com/libgit2/git2go/v30"
	"github.com/pkg/errors"
)

// TODO: Is an in memory implementation possible?

// Repository ...
type Repository struct {
	urs  string
	path string
	opts *RepositoryOpts

	key        *keys.EdX25519Key
	ks         *keys.Store
	publicKey  string
	privateKey string

	host string

	repo *git.Repository
}

// RepositoryOpts are options for repository.
type RepositoryOpts struct {
	GitUser string
}

// NewRepository ...
func NewRepository(urs string, host string, path string, key *keys.EdX25519Key, opts *RepositoryOpts) (*Repository, error) {
	if opts == nil {
		opts = &RepositoryOpts{}
	}
	if opts.GitUser == "" {
		opts.GitUser = "git"
	}

	privateKey, err := key.EncodeToSSH(nil)
	if err != nil {
		return nil, err
	}
	publicKey := key.PublicKey().EncodeToSSHAuthorized()
	logger.Debugf("Git url: %s", urs)
	logger.Debugf("Git public key: %s", publicKey)
	logger.Debugf("Git user: %s", opts.GitUser)

	ks := keys.NewMemStore(true)
	if err := ks.Save(key); err != nil {
		return nil, err
	}

	return &Repository{
		path:       path,
		key:        key,
		ks:         ks,
		opts:       opts,
		urs:        urs,
		host:       host,
		publicKey:  string(publicKey),
		privateKey: string(privateKey),
	}, nil
}

func (r *Repository) credentialsCallback(url string, usernameFromURL string, allowedTypes git.CredType) (*git.Cred, error) {
	cred, err := git.NewCredSshKeyFromMemory(r.opts.GitUser, r.publicKey, r.privateKey, "")
	if err != nil {
		return nil, err
	}
	return cred, nil
}

func (r *Repository) newRemoteCallbacks() git.RemoteCallbacks {
	return git.RemoteCallbacks{
		CredentialsCallback: r.credentialsCallback,
		CertificateCheckCallback: func(cert *git.Certificate, valid bool, hostname string) git.ErrorCode {
			logger.Debugf("Certificate check %t %s", valid, hostname)
			// if !valid {
			// 	return git.ErrCertificate
			// }
			if hostname != r.host {
				return git.ErrCertificate
			}
			return git.ErrOk
		},
	}
}

// Open repository.
func (r *Repository) Open() error {
	if r.repo != nil {
		return errors.Errorf("already open")
	}

	logger.Debugf("Open repo: %s", r.path)
	if _, err := os.Stat(r.path); err == nil {
		repo, err := git.OpenRepository(r.path)
		if err != nil {
			return errors.Wrapf(err, "failed to open repository")
		}
		r.repo = repo
	} else {
		logger.Debugf("Clone repo: %s", r.urs)
		opts := &git.CloneOptions{}
		opts.FetchOptions = &git.FetchOptions{
			RemoteCallbacks: r.newRemoteCallbacks(),
		}
		repo, err := git.Clone(r.urs, r.path, opts)
		if err != nil {
			return errors.Wrapf(err, "failed to clone repository")
		}
		r.repo = repo
	}

	logger.Debugf("Repo: %s", r.repo.Path())

	// if remote == nil {
	// remote, err := r.repo.Remotes.Create("origin", r.repo.Path())
	// if err != nil {
	// 	return errors.Wrap(err, "failed to create origin")
	// }
	// r.remote = remote
	// }
	// if err := r.remote.ConnectFetch(&rcb, nil, nil); err != nil {
	// 	return errors.Wrapf(err, "failed to connect/fetch")
	// }

	// heads, err := r.remote.Ls()
	// if err != nil {
	// 	return errors.Wrapf(err, "failed to ls remote")
	// }
	// logger.Debugf("Heads: %v", heads)

	return nil
}

// Close repo.
func (r *Repository) Close() {
	if r.repo != nil {
		r.repo.Free()
		r.repo = nil
	}
}

// Pull changes.
func (r *Repository) Pull() error {
	if r.repo == nil {
		return errors.Errorf("not open")
	}

	remote, err := r.repo.Remotes.Lookup("origin")
	if err != nil {
		return errors.Wrap(err, "failed to lookup origin")
	}
	if remote == nil {
		return errors.Errorf("no remote origin")
	}

	logger.Debugf("Fetch refspecs")
	refspecs, err := remote.FetchRefspecs()
	if err != nil {
		return err
	}
	logger.Debugf("Fetch refspecs: %+v", refspecs)

	logger.Debugf("Fetch")
	opts := &git.FetchOptions{RemoteCallbacks: r.newRemoteCallbacks()}
	if err := remote.Fetch(refspecs, opts, "Fetching remote"); err != nil {
		return errors.Wrap(err, "failed to push")
	}

	remoteBranch, err := r.repo.References.Lookup("refs/remotes/origin/master")
	if err != nil {
		return err
	}

	remoteBranchID := remoteBranch.Target()
	annotatedCommit, err := r.repo.AnnotatedCommitFromRef(remoteBranch)
	if err != nil {
		return err
	}

	analysis, _, err := r.repo.MergeAnalysis([]*git.AnnotatedCommit{annotatedCommit})
	if err != nil {
		return err
	}

	// Get repo head
	head, err := r.repo.Head()
	if err != nil {
		return err
	}

	logger.Debugf("Git merge analysis: %s", mergeAnalysisDescription(analysis))

	if isMergeAnalysis(analysis, git.MergeAnalysisUpToDate) {
		logger.Debugf("Up to date")
		// Up to date
		return nil
	}

	if isMergeAnalysis(analysis, git.MergeAnalysisFastForward) {
		logger.Debugf("Fast forward")

		target := remoteBranch.Target()

		commit, err := r.repo.LookupCommit(target)
		if err != nil {
			return err
		}

		commitTree, err := commit.Tree()
		if err != nil {
			return err
		}

		logger.Debugf("Checkout tree")
		err = r.repo.CheckoutTree(commitTree, &git.CheckoutOpts{Strategy: git.CheckoutSafe})
		if err != nil {
			return err
		}

		logger.Debugf("Set target")
		if _, err := head.SetTarget(target, ""); err != nil {
			return err
		}

		return nil
	}

	if isMergeAnalysis(analysis, git.MergeAnalysisNormal) {
		logger.Debugf("Normal")

		// Just merge changes
		if err := r.repo.Merge([]*git.AnnotatedCommit{annotatedCommit}, nil, nil); err != nil {
			return err
		}
		// Check for conflicts
		index, err := r.repo.Index()
		if err != nil {
			return err
		}

		if index.HasConflicts() {
			return errors.New("git conflicts")
		}

		// Get Write Tree
		treeID, err := index.WriteTree()
		if err != nil {
			return err
		}

		tree, err := r.repo.LookupTree(treeID)
		if err != nil {
			return err
		}

		localCommit, err := r.repo.LookupCommit(head.Target())
		if err != nil {
			return err
		}

		remoteCommit, err := r.repo.LookupCommit(remoteBranchID)
		if err != nil {
			return err
		}

		sig := r.signature()
		commitID, err := r.repo.CreateCommit("HEAD", sig, sig, "", tree, localCommit, remoteCommit)
		if err != nil {
			return err
		}
		logger.Debugf("Merge commit: %s", commitID)

		// Clean up
		if err := r.repo.StateCleanup(); err != nil {
			return err
		}

		return nil
	}

	return errors.Errorf("unhandled merge analysis: %s", mergeAnalysisDescription(analysis))
}

func isMergeAnalysis(m1 git.MergeAnalysis, m2 git.MergeAnalysis) bool {
	return m1&m2 != 0
}

func mergeAnalysisDescription(m git.MergeAnalysis) string {
	descs := []string{}
	if isMergeAnalysis(m, git.MergeAnalysisUpToDate) {
		descs = append(descs, "up-to-date")
	}
	if isMergeAnalysis(m, git.MergeAnalysisFastForward) {
		descs = append(descs, "fast-forward")
	}
	if isMergeAnalysis(m, git.MergeAnalysisNormal) {
		descs = append(descs, "normal")
	}
	if isMergeAnalysis(m, git.MergeAnalysisUnborn) {
		descs = append(descs, "unborn")
	}
	if isMergeAnalysis(m, git.MergeAnalysisNone) {
		descs = append(descs, "none")
	}
	if len(descs) == 0 {
		return "unknown"
	}

	return strings.Join(descs, ", ")
}

// Push changes.
func (r *Repository) Push() error {
	if r.repo == nil {
		return errors.Errorf("not open")
	}

	remote, err := r.repo.Remotes.Lookup("origin")
	if err != nil {
		return errors.Wrap(err, "failed to lookup origin")
	}
	if remote == nil {
		return errors.Errorf("no remote origin")
	}

	opts := &git.PushOptions{RemoteCallbacks: r.newRemoteCallbacks()}
	if err := remote.Push([]string{"refs/heads/master"}, opts); err != nil {
		return errors.Wrap(err, "failed to push")
	}
	return nil
}

func (r *Repository) signature() *git.Signature {
	return &git.Signature{
		Name:  "keys.pub",
		Email: "git@keys.pub",
		When:  time.Now(),
	}
}

// Add ...
func (r *Repository) Add(item *keyring.Item) error {
	if r.repo == nil {
		return errors.Errorf("not open")
	}

	idx, err := r.repo.Index()
	if err != nil {
		return err
	}

	name := item.ID
	path := filepath.Join(r.path, name)

	encrypted, err := encryptItem(item, r.key)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(path, encrypted, 0600); err != nil {
		return err
	}

	if err := idx.AddByPath(name); err != nil {
		return err
	}
	if err := idx.Write(); err != nil {
		return err
	}
	treeID, err := idx.WriteTree()
	if err != nil {
		return err
	}

	logger.Debugf("Add %s", name)
	message := fmt.Sprintf("Add %s\n", name)
	if err := r.createCommit(treeID, message); err != nil {
		return err
	}

	return nil
}

// Delete ...
func (r *Repository) Delete(id string) error {
	if r.repo == nil {
		return errors.Errorf("not open")
	}

	name := id
	path := filepath.Join(r.path, name)

	if err := os.Remove(path); err != nil {
		return err
	}

	idx, err := r.repo.Index()
	if err != nil {
		return err
	}

	if err := idx.RemoveByPath(name); err != nil {
		return err
	}
	if err := idx.Write(); err != nil {
		return err
	}
	treeID, err := idx.WriteTree()
	if err != nil {
		return err
	}

	logger.Debugf("Delete item %s", name)
	message := fmt.Sprintf("Delete %s\n", name)
	if err := r.createCommit(treeID, message); err != nil {
		return err
	}
	return nil
}

func (r *Repository) createCommit(treeID *git.Oid, message string) error {
	sig := r.signature()

	currentBranch, err := r.repo.Head()
	if err != nil {
		return err
	}
	currentTip, err := r.repo.LookupCommit(currentBranch.Target())
	if err != nil {
		return err
	}

	tree, err := r.repo.LookupTree(treeID)
	if err != nil {
		return err
	}
	commitID, err := r.repo.CreateCommit("HEAD", sig, sig, message, tree, currentTip)
	if err != nil {
		return err
	}
	logger.Debugf("Commit %s", commitID)
	return nil
}
