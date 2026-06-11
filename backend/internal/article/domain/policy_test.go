package domain_test

import (
	"testing"

	"github.com/yourorg/goapp/internal/article/domain"
	"github.com/yourorg/goapp/internal/platform/authz"
)

func TestPolicy(t *testing.T) {
	owner := &authz.Actor{ID: 1, Role: authz.RoleUser}
	stranger := &authz.Actor{ID: 2, Role: authz.RoleUser}
	admin := &authz.Actor{ID: 3, Role: authz.RoleAdmin}

	draft := domain.Article{ID: 10, AuthorID: 1, Status: domain.StatusDraft}
	published := domain.Article{ID: 11, AuthorID: 1, Status: domain.StatusPublished}

	tests := []struct {
		name       string
		actor      *authz.Actor
		record     domain.Article
		canView    bool
		canEdit    bool
		canPublish bool
	}{
		{"owner on draft", owner, draft, true, true, true},
		{"stranger on draft", stranger, draft, false, false, false},
		{"stranger on published", stranger, published, true, false, false},
		{"admin on draft", admin, draft, true, true, true},
		{"anonymous on published", nil, published, true, false, false},
		{"anonymous on draft", nil, draft, false, false, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := domain.NewPolicy(tc.actor, tc.record)
			if got := p.CanView(); got != tc.canView {
				t.Errorf("CanView = %v, want %v", got, tc.canView)
			}
			if got := p.CanEdit(); got != tc.canEdit {
				t.Errorf("CanEdit = %v, want %v", got, tc.canEdit)
			}
			if got := p.CanPublish(); got != tc.canPublish {
				t.Errorf("CanPublish = %v, want %v", got, tc.canPublish)
			}
		})
	}
}
