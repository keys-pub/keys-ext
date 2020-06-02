package git

import (
	"strings"
	"time"

	git "github.com/keys-pub/git2go"
	"github.com/pkg/errors"
)

func (r *Repository) merge() error {
	logger.Debugf("Merge...")
	if r.repo == nil {
		return errors.Errorf("not open")
	}

	repo, err := git.OpenRepository(r.path)
	if err != nil {
		return errors.Wrapf(err, "failed to open repo (git2go)")
	}
	defer repo.Free()

	remoteBranch, err := repo.References.Lookup("refs/remotes/origin/master")
	if err != nil {
		return errors.Wrapf(err, "failed to lookup refs/remotes/origin/master")
	}

	remoteBranchID := remoteBranch.Target()
	annotatedCommit, err := repo.AnnotatedCommitFromRef(remoteBranch)
	if err != nil {
		return errors.Wrapf(err, "failed to lookup remote branch commit")
	}

	analysis, _, err := repo.MergeAnalysis([]*git.AnnotatedCommit{annotatedCommit})
	if err != nil {
		return errors.Wrapf(err, "failed merge analysis")
	}

	// Get repo head
	head, err := repo.Head()
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

		commit, err := repo.LookupCommit(target)
		if err != nil {
			return err
		}

		commitTree, err := commit.Tree()
		if err != nil {
			return err
		}

		logger.Debugf("Checkout tree")
		err = repo.CheckoutTree(commitTree, &git.CheckoutOpts{Strategy: git.CheckoutSafe})
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

		if err := repo.Merge([]*git.AnnotatedCommit{annotatedCommit}, nil, checkoutOpts); err != nil {
			return err
		}
		// Check for conflicts
		idx, err := repo.Index()
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

		tree, err := repo.LookupTree(treeID)
		if err != nil {
			return err
		}

		localCommit, err := repo.LookupCommit(head.Target())
		if err != nil {
			return err
		}

		remoteCommit, err := repo.LookupCommit(remoteBranchID)
		if err != nil {
			return err
		}

		sig := &git.Signature{
			Name:  "keys.pub",
			Email: "git@keys.pub",
			When:  time.Now(),
		}
		message := "Merge"
		commitID, err := repo.CreateCommit("HEAD", sig, sig, message, tree, localCommit, remoteCommit)
		if err != nil {
			return err
		}
		logger.Debugf("Merge commit: %s", commitID)

		// Clean up
		if err := repo.StateCleanup(); err != nil {
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
