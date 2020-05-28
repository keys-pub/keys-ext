package git

import (
	"fmt"
	"os"
	"strings"
	"time"

	git "github.com/keys-pub/git2go"
	"github.com/keys-pub/keys"
	"github.com/pkg/errors"
)

// Repository ...
type Repository struct {
	publicKey  string
	privateKey string
	gitUser    string

	repo *git.Repository
	path string
}

// NewRepository ...
func NewRepository() *Repository {
	return &Repository{}
}

// Path to repo.
func (r *Repository) Path() string {
	return r.path
}

func (r *Repository) credentialsCallback(url string, usernameFromURL string, allowedTypes git.CredType) (*git.Cred, error) {
	if r.privateKey == "" || r.publicKey == "" {
		return nil, errors.Errorf("no ssh key set")
	}
	gitUser := r.gitUser
	if gitUser == "" {
		gitUser = "git"
	}

	cred, err := git.NewCredSshKeyFromMemory(gitUser, r.publicKey, r.privateKey, "")
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
			// TODO: Check cert
			// if !valid {
			// 	return git.ErrCertificate
			// }
			// if hostname != r.host {
			// 	return git.ErrCertificate
			// }
			return git.ErrOk
		},
	}
}

func pathExists(path string) (bool, error) {
	if _, err := os.Stat(path); err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	} else {
		return false, err
	}
}

// Open repository.
func (r *Repository) Open(path string) error {
	if r.repo != nil {
		return errors.Errorf("already open")
	}

	logger.Debugf("Open repo: %s", path)
	exist, err := pathExists(path)
	if err != nil {
		return err
	}
	if !exist {
		return errors.Errorf("path doesn't exist: %s", r.path)
	}

	repo, err := git.OpenRepository(path)
	if err != nil {
		return errors.Wrapf(err, "failed to open repository")
	}
	r.repo = repo
	r.path = path
	logger.Debugf("Opened repo: %s", r.repo.Path())
	return nil
}

// SetKey sets the ssh key.
func (r *Repository) SetKey(key *keys.EdX25519Key) error {
	privateKey, err := key.EncodeToSSH(nil)
	if err != nil {
		return err
	}
	publicKey := key.PublicKey().EncodeToSSHAuthorized()
	logger.Debugf("Git public key: %s", publicKey)
	r.privateKey = string(privateKey)
	r.publicKey = string(publicKey)
	return nil
}

// Clone repository.
func (r *Repository) Clone(urs string, path string) error {
	if r.repo != nil {
		return errors.Errorf("already open")
	}

	logger.Debugf("Clone repo: %s", urs)
	opts := &git.CloneOptions{}
	opts.FetchOptions = &git.FetchOptions{
		RemoteCallbacks: r.newRemoteCallbacks(),
	}
	repo, err := git.Clone(urs, path, opts)
	if err != nil {
		return errors.Wrapf(err, "failed to clone repository")
	}
	r.repo = repo
	r.path = path
	logger.Debugf("Cloned repo: %s", r.repo.Path())
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
		logger.Debugf("Merge analysis: Up to date")
		// Up to date
		return nil
	}

	if isMergeAnalysis(analysis, git.MergeAnalysisFastForward) {
		logger.Debugf("Merge analysis: Fast forward")

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
		logger.Debugf("Merge analysis: Normal")

		checkoutOpts := &git.CheckoutOpts{
			Strategy: git.CheckoutSafe | git.CheckoutRecreateMissing | git.CheckoutAllowConflicts | git.CheckoutUseOurs,
		}

		if err := r.repo.Merge([]*git.AnnotatedCommit{annotatedCommit}, nil, checkoutOpts); err != nil {
			return err
		}
		// Check for conflicts
		idx, err := r.repo.Index()
		if err != nil {
			return err
		}

		if idx.HasConflicts() {
			if err := idx.AddAll([]string{}, git.IndexAddDefault, nil); err != nil {
				return err
			}
			if err := idx.Write(); err != nil {
				return err
			}
		}

		// Get Write Tree
		treeID, err := idx.WriteTree()
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
		message := "Merge"
		commitID, err := r.repo.CreateCommit("HEAD", sig, sig, message, tree, localCommit, remoteCommit)
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

func (r *Repository) add(name string) error {
	if r.repo == nil {
		return errors.Errorf("not open")
	}

	idx, err := r.repo.Index()
	if err != nil {
		return errors.Wrapf(err, "failed to get index")
	}

	if err := idx.AddByPath(name); err != nil {
		return errors.Wrapf(err, "failed to add by path")
	}
	if err := idx.Write(); err != nil {
		return errors.Wrapf(err, "failed to write add")
	}
	treeID, err := idx.WriteTree()
	if err != nil {
		return errors.Wrapf(err, "failed to write tree")
	}

	logger.Debugf("Add %s", name)
	message := fmt.Sprintf("Add %s\n", name)
	if err := r.createCommit(treeID, message); err != nil {
		return errors.Wrapf(err, "failed to create commit")
	}

	return nil
}

func (r *Repository) delete(name string) error {
	if r.repo == nil {
		return errors.Errorf("not open")
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

	logger.Debugf("Delete %s", name)
	message := fmt.Sprintf("Delete %s\n", name)
	if err := r.createCommit(treeID, message); err != nil {
		return err
	}
	return nil
}

func (r *Repository) createCommit(treeID *git.Oid, message string) error {
	unborn := false
	currentBranch, err := r.repo.Head()
	if err != nil {
		if ErrIsCode(err, ErrUnbornBranch) {
			// No commits yet
			unborn = true
		} else {
			return errors.Wrapf(err, "failed to get head")
		}
	}

	tree, err := r.repo.LookupTree(treeID)
	if err != nil {
		return errors.Wrapf(err, "failed to lookup tree")
	}

	sig := r.signature()
	if unborn {
		// If unborn then create commit with no parent, this happens on a bare
		// repository.
		commitID, err := r.repo.CreateCommit("HEAD", sig, sig, message, tree)
		if err != nil {
			return err
		}
		logger.Debugf("Commit (first) %s", commitID)
	} else {
		currentTip, err := r.repo.LookupCommit(currentBranch.Target())
		if err != nil {
			return errors.Wrapf(err, "failed to lookup commit")
		}

		commitID, err := r.repo.CreateCommit("HEAD", sig, sig, message, tree, currentTip)
		if err != nil {
			return err
		}
		logger.Debugf("Commit %s", commitID)
	}
	return nil
}
