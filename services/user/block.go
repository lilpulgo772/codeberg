// Copyright 2023 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package user

import (
	"context"

	"code.gitea.io/gitea/models/db"
	repo_model "code.gitea.io/gitea/models/repo"
	user_model "code.gitea.io/gitea/models/user"
)

// BlockUser adds a blocked user entry for userID to block blockID.
// TODO: Figure out if instance admins should be immune to blocking.
// TODO: Add more mechanism like removing blocked user as collaborator on
// repositories where the user is an owner.
func BlockUser(ctx context.Context, userID, blockID int64) error {
	if userID == blockID || user_model.IsBlocked(ctx, userID, blockID) {
		return nil
	}

	ctx, committer, err := db.TxContext(ctx)
	if err != nil {
		return err
	}
	defer committer.Close()

	// Add the blocked user entry.
	_, err = db.GetEngine(ctx).Insert(&user_model.BlockedUser{UserID: userID, BlockID: blockID})
	if err != nil {
		return err
	}

	// Unfollow the user from the block's perspective.
	err = user_model.UnfollowUser(ctx, blockID, userID)
	if err != nil {
		return err
	}

	// Unfollow the user from the doer's perspective.
	err = user_model.UnfollowUser(ctx, userID, blockID)
	if err != nil {
		return err
	}

	// Blocked user unwatch all repository owned by the doer.
	repoIDs, err := repo_model.GetWatchedRepoIDsOwnedBy(ctx, blockID, userID)
	if err != nil {
		return err
	}

	err = repo_model.UnwatchRepos(ctx, blockID, repoIDs)
	if err != nil {
		return err
	}

	// Remove blocked user as collaborator from repositories the user owns as an
	// individual.
	collabsID, err := repo_model.GetCollaboratorWithUser(ctx, userID, blockID)
	if err != nil {
		return err
	}

	_, err = db.GetEngine(ctx).In("id", collabsID).Delete(&repo_model.Collaboration{})
	if err != nil {
		return err
	}

	return committer.Commit()
}
