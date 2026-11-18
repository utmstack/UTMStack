package upgrade

import (
	"strings"

	"github.com/threatwinds/go-sdk/catcher"
	"gorm.io/gorm"
)

type v11User struct {
	ID           int64
	Login        string
	PasswordHash string
	FirstName    *string
	LastName     *string
	Email        *string
	ImageURL     *string
	Activated    bool
	LangKey      *string
}

// v11UserUUID derives the v12 user id from the v11 numeric surrogate id.
// Deriving from the id (not the login) keeps the user and user_role
// converters independent: jhi_user_authority only carries the numeric user_id.
func v11UserUUID(v11ID int64) string {
	return id5(userNamespace, v11ID)
}

func mapRoleID(authority string) string {
	switch authority {
	case "ROLE_ADMIN":
		return roleAdminID
	case "ROLE_USER":
		return roleAnalystID
	default:
		return roleViewerID
	}
}

func convertUsers(db *gorm.DB) error {
	if !tableExists(db, "jhi_user") {
		logSkipTable("jhi_user")
		return nil
	}
	var src []v11User
	if err := db.Raw(`SELECT id, login, password_hash, first_name, last_name,
		email, image_url, activated, lang_key FROM jhi_user ORDER BY id`).Scan(&src).Error; err != nil {
		return err
	}

	inserted := 0
	if err := db.Transaction(func(tx *gorm.DB) error {
		for _, u := range src {
			email := strFrom(u.Email)
			if email == "" {
				email = u.Login
			}
			name := strings.TrimSpace(strFrom(u.FirstName) + " " + strFrom(u.LastName))
			if name == "" {
				name = u.Login
			}
			status := "inactive"
			if u.Activated {
				status = "active"
			}

			if err := tx.Exec(
				`INSERT INTO "user" (id, tenant_id, email, name, password_hash, status, lang_key, image_url)
				 VALUES (?, ?, ?, ?, ?, ?, ?, ?)
				 ON CONFLICT (id) DO UPDATE SET
					tenant_id = EXCLUDED.tenant_id,
					email = EXCLUDED.email,
					name = EXCLUDED.name,
					password_hash = EXCLUDED.password_hash,
					status = EXCLUDED.status,
					lang_key = EXCLUDED.lang_key,
					image_url = EXCLUDED.image_url`,
				v11UserUUID(u.ID), defaultTenantID, email, name, u.PasswordHash, status, strFrom(u.LangKey), strFrom(u.ImageURL),
			).Error; err != nil {
				return err
			}
			inserted++
		}
		return nil
	}); err != nil {
		return err
	}

	roles, err := convertUserRoles(db)
	if err != nil {
		return err
	}

	catcher.Info("v11 upgrade: users converted", map[string]any{"users": inserted, "roles": roles})
	return nil
}

type v11UserAuthority struct {
	UserID        int64
	AuthorityName string
}

func convertUserRoles(db *gorm.DB) (int, error) {
	if !tableExists(db, "jhi_user_authority") {
		logSkipTable("jhi_user_authority")
		return 0, nil
	}

	var src []v11UserAuthority
	if err := db.Raw(`SELECT user_id, authority_name FROM jhi_user_authority ORDER BY user_id`).Scan(&src).Error; err != nil {
		return 0, err
	}

	inserted := 0
	if err := db.Transaction(func(tx *gorm.DB) error {
		for _, a := range src {
			if err := tx.Exec(
				`INSERT INTO user_role (user_id, role_id, tenant_id)
				 VALUES (?, ?, ?)
				 ON CONFLICT (user_id, role_id) DO NOTHING`,
				v11UserUUID(a.UserID), mapRoleID(a.AuthorityName), defaultTenantID,
			).Error; err != nil {
				return err
			}
			inserted++
		}
		return nil
	}); err != nil {
		return 0, err
	}
	return inserted, nil
}

func strFrom(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
