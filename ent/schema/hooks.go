package schema

import (
	"context"
	"errors"
	"fmt"

	"github.com/a8m/ent-sync-example/ent"
	"github.com/a8m/ent-sync-example/ent/hook"
)

// EnsureImageExists ensures the avatar_url points
// to a real object in the bucket.
func EnsureImageExists() ent.Hook {
	hk := func(next ent.Mutator) ent.Mutator {
		return hook.UserFunc(func(ctx context.Context, m *ent.UserMutation) (ent.Value, error) {
			avatar, exists := m.AvatarURL()
			if !exists {
				return nil, errors.New("avatar field is missing")
			}
			switch exists, err := m.Bucket.Exists(ctx, avatar); {
			case err != nil:
				return nil, fmt.Errorf("check key existence: %w", err)
			case !exists:
				return nil, fmt.Errorf("key %q does not exist in the bucket", avatar)
			default:
				return next.Mutate(ctx, m)
			}
		})
	}
	// Limit the hook only to "Create" operations.
	return hook.On(hk, ent.OpCreate)
}

// DeleteOrphans cascades the user deletion to the bucket.
// Hence, when a user is deleted, its avatar image is deleted
// as well.
func DeleteOrphans() ent.Hook {
	hk := func(next ent.Mutator) ent.Mutator {
		return hook.UserFunc(func(ctx context.Context, m *ent.UserMutation) (ent.Value, error) {
			id, exists := m.ID()
			if !exists {
				return nil, errors.New("id field is missing")
			}
			u, err := m.Client().User.Get(ctx, id)
			if err != nil {
				return nil, fmt.Errorf("getting deleted user: %w", err)
			}
			if err := m.Bucket.Delete(ctx, u.AvatarURL); err != nil {
				return nil, fmt.Errorf("deleting user avatar from bucket: %w", err)
			}
			return next.Mutate(ctx, m)
		})
	}
	// Limit the hook only to "DeleteOne" operations.
	// For example,
	//
	//	client.User.DeleteOne(usr).Exec(ctx)
	//	client.User.DeleteOneID(id).Exec(ctx)
	//
	return hook.On(hk, ent.OpDeleteOne)
}
