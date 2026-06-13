package handlers

import (
	"context"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/fresp/Statora/internal/models"
	"github.com/fresp/Statora/internal/repository"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	validThemePresets = map[string]struct{}{
		"default":  {},
		"ocean":    {},
		"graphite": {},
		"ember":    {},
		"frost":    {},
	}
	colorHexPattern = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)
)

func isValidThemePreset(value string) bool {
	if value == "" {
		return true
	}
	// hanya allow safe string (prevent injection/path traversal)
	matched, _ := regexp.MatchString(`^[a-z0-9-_]+$`, value)
	return matched
}

func isValidColorHex(value string) bool {
	return colorHexPattern.MatchString(value)
}

func isValidURLOrEmpty(value string) bool {
	if value == "" {
		return true
	}
	parsed, err := url.ParseRequestURI(value)
	if err != nil {
		return false
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return false
	}
	return parsed.Scheme == "http" || parsed.Scheme == "https"
}

func normalizeThemePreset(value string) string {
	normalized := strings.TrimSpace(strings.ToLower(value))
	return strings.TrimSuffix(normalized, ".css")
}

type statusPageSSOResponse struct {
	Enabled      bool   `json:"enabled"`
	Provider     string `json:"provider"`
	Issuer       string `json:"issuer"`
	Audience     string `json:"audience"`
	Algorithm    string `json:"algorithm"`
	PublicKeyPEM string `json:"publicKeyPem"`
	HasSecret    bool   `json:"hasSecret"`
}

type adminStatusPageSettingsResponse struct {
	Head      models.StatusPageHeadSettings     `json:"head"`
	Branding  models.StatusPageBrandingSettings `json:"branding"`
	Theme     models.StatusPageThemeSettings    `json:"theme"`
	Layout    models.StatusPageLayoutSettings   `json:"layout"`
	Footer    models.StatusPageFooterSettings   `json:"footer"`
	SSO       statusPageSSOResponse             `json:"sso"`
	CustomCSS string                            `json:"customCss"`
	UpdatedAt time.Time                         `json:"updatedAt"`
	CreatedAt time.Time                         `json:"createdAt"`
}

type statusPageSettingsPatchRequest struct {
	Head *struct {
		Title       *string            `json:"title"`
		Description *string            `json:"description"`
		Keywords    *string            `json:"keywords"`
		FaviconURL  *string            `json:"faviconUrl"`
		MetaTags    *map[string]string `json:"metaTags"`
	} `json:"head"`
	Branding *struct {
		SiteName           *string `json:"siteName"`
		LogoURL            *string `json:"logoUrl"`
		BackgroundImageURL *string `json:"backgroundImageUrl"`
		HeroImageURL       *string `json:"heroImageUrl"`
	} `json:"branding"`
	Theme *struct {
		Preset *string `json:"preset"`
		Mode   *string `json:"mode"`
		Light  *struct {
			PrimaryColor    *string `json:"primaryColor"`
			BackgroundColor *string `json:"backgroundColor"`
			TextColor       *string `json:"textColor"`
			AccentColor     *string `json:"accentColor"`
		} `json:"light"`
		Dark *struct {
			PrimaryColor    *string `json:"primaryColor"`
			BackgroundColor *string `json:"backgroundColor"`
			TextColor       *string `json:"textColor"`
			AccentColor     *string `json:"accentColor"`
		} `json:"dark"`
		Typography *struct {
			FontFamily *string `json:"fontFamily"`
			FontScale  *string `json:"fontScale"`
		} `json:"typography"`
	} `json:"theme"`
	Layout *struct {
		Variant *string `json:"variant"`
	} `json:"layout"`
	Footer *struct {
		Text          *string `json:"text"`
		ShowPoweredBy *bool   `json:"showPoweredBy"`
	} `json:"footer"`
	SSO *struct {
		Enabled      *bool   `json:"enabled"`
		Provider     *string `json:"provider"`
		Issuer       *string `json:"issuer"`
		Audience     *string `json:"audience"`
		Algorithm    *string `json:"algorithm"`
		SharedSecret *string `json:"sharedSecret"`
		PublicKeyPEM *string `json:"publicKeyPem"`
	} `json:"sso"`
	CustomCSS *string `json:"customCss"`
}

func settingsCollection(db *mongo.Database) *mongo.Collection {
	return db.Collection("settings")
}

func fetchOrCreateStatusPageSettings(ctx context.Context, db *mongo.Database) (models.StatusPageSettings, error) {
	var settings models.StatusPageSettings
	err := settingsCollection(db).FindOne(ctx, bson.M{"key": models.StatusPageSettingsKey}).Decode(&settings)
	if err == nil {
		return settings, nil
	}
	if err != mongo.ErrNoDocuments {
		return models.StatusPageSettings{}, err
	}

	defaultSettings := models.DefaultStatusPageSettings()
	if _, insertErr := settingsCollection(db).InsertOne(ctx, defaultSettings); insertErr != nil {
		return models.StatusPageSettings{}, insertErr
	}
	return defaultSettings, nil
}

func GetPublicStatusPageSettings(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		settings, err := fetchOrCreateStatusPageSettings(ctx, db)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, settings)
	}
}

func GetAdminStatusPageSettings(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		settings, err := fetchOrCreateStatusPageSettings(ctx, db)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, buildAdminStatusPageSettingsResponse(settings))
	}
}

func buildAdminStatusPageSettingsResponse(settings models.StatusPageSettings) adminStatusPageSettingsResponse {
	return adminStatusPageSettingsResponse{
		Head:     settings.Head,
		Branding: settings.Branding,
		Theme:    settings.Theme,
		Layout:   settings.Layout,
		Footer:   settings.Footer,
		SSO: statusPageSSOResponse{
			Enabled:      settings.SSO.Enabled,
			Provider:     settings.SSO.Provider,
			Issuer:       settings.SSO.Issuer,
			Audience:     settings.SSO.Audience,
			Algorithm:    settings.SSO.Algorithm,
			PublicKeyPEM: settings.SSO.PublicKeyPEM,
			HasSecret:    strings.TrimSpace(settings.SSO.SharedSecret) != "",
		},
		CustomCSS: settings.CustomCSS,
		UpdatedAt: settings.UpdatedAt,
		CreatedAt: settings.CreatedAt,
	}
}

func UpdateStatusPageSettings(db *mongo.Database, hub *Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req statusPageSettingsPatchRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		settingsRepo := repository.NewMongoSettingsRepository(db)
		current, err := fetchOrCreateStatusPageSettings(ctx, db)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		set := bson.M{"updatedAt": time.Now()}

		if req.Head != nil {
			if req.Head.Title != nil {
				set["head.title"] = *req.Head.Title
			}
			if req.Head.Description != nil {
				set["head.description"] = *req.Head.Description
			}
			if req.Head.Keywords != nil {
				set["head.keywords"] = *req.Head.Keywords
			}
			if req.Head.FaviconURL != nil {
				if !isValidURLOrEmpty(*req.Head.FaviconURL) {
					c.JSON(http.StatusBadRequest, gin.H{"error": "head.faviconUrl must be a valid http(s) URL or empty"})
					return
				}
				set["head.faviconUrl"] = *req.Head.FaviconURL
			}
			if req.Head.MetaTags != nil {
				set["head.metaTags"] = *req.Head.MetaTags
			}
		}

		if req.Branding != nil {
			if req.Branding.SiteName != nil {
				set["branding.siteName"] = *req.Branding.SiteName
			}
			if req.Branding.LogoURL != nil {
				if !isValidURLOrEmpty(*req.Branding.LogoURL) {
					c.JSON(http.StatusBadRequest, gin.H{"error": "branding.logoUrl must be a valid http(s) URL or empty"})
					return
				}
				set["branding.logoUrl"] = *req.Branding.LogoURL
			}
			if req.Branding.BackgroundImageURL != nil {
				if !isValidURLOrEmpty(*req.Branding.BackgroundImageURL) {
					c.JSON(http.StatusBadRequest, gin.H{"error": "branding.backgroundImageUrl must be a valid http(s) URL or empty"})
					return
				}
				set["branding.backgroundImageUrl"] = *req.Branding.BackgroundImageURL
			}
			if req.Branding.HeroImageURL != nil {
				if !isValidURLOrEmpty(*req.Branding.HeroImageURL) {
					c.JSON(http.StatusBadRequest, gin.H{"error": "branding.heroImageUrl must be a valid http(s) URL or empty"})
					return
				}
				set["branding.heroImageUrl"] = *req.Branding.HeroImageURL
			}
		}

		if req.Theme != nil {
			if req.Theme.Preset != nil {
				normalizedPreset := normalizeThemePreset(*req.Theme.Preset)

				if !isValidThemePreset(normalizedPreset) {
					c.JSON(http.StatusBadRequest, gin.H{"error": "invalid theme preset format"})
					return
				}
				set["theme.preset"] = normalizedPreset
			}
			if req.Theme.Mode != nil {
				if *req.Theme.Mode != "light" && *req.Theme.Mode != "dark" && *req.Theme.Mode != "system" {
					c.JSON(http.StatusBadRequest, gin.H{"error": "theme.mode must be one of: light, dark, system"})
					return
				}
				set["theme.mode"] = *req.Theme.Mode
			}
			if req.Theme.Light != nil {
				if req.Theme.Light.PrimaryColor != nil {
					if !isValidColorHex(*req.Theme.Light.PrimaryColor) {
						c.JSON(http.StatusBadRequest, gin.H{"error": "theme.light.primaryColor must be a valid hex color (#RRGGBB)"})
						return
					}
					set["theme.light.primaryColor"] = *req.Theme.Light.PrimaryColor
				}
				if req.Theme.Light.BackgroundColor != nil {
					if !isValidColorHex(*req.Theme.Light.BackgroundColor) {
						c.JSON(http.StatusBadRequest, gin.H{"error": "theme.light.backgroundColor must be a valid hex color (#RRGGBB)"})
						return
					}
					set["theme.light.backgroundColor"] = *req.Theme.Light.BackgroundColor
				}
				if req.Theme.Light.TextColor != nil {
					if !isValidColorHex(*req.Theme.Light.TextColor) {
						c.JSON(http.StatusBadRequest, gin.H{"error": "theme.light.textColor must be a valid hex color (#RRGGBB)"})
						return
					}
					set["theme.light.textColor"] = *req.Theme.Light.TextColor
				}
				if req.Theme.Light.AccentColor != nil {
					if !isValidColorHex(*req.Theme.Light.AccentColor) {
						c.JSON(http.StatusBadRequest, gin.H{"error": "theme.light.accentColor must be a valid hex color (#RRGGBB)"})
						return
					}
					set["theme.light.accentColor"] = *req.Theme.Light.AccentColor
				}
			}
			if req.Theme.Dark != nil {
				if req.Theme.Dark.PrimaryColor != nil {
					if !isValidColorHex(*req.Theme.Dark.PrimaryColor) {
						c.JSON(http.StatusBadRequest, gin.H{"error": "theme.dark.primaryColor must be a valid hex color (#RRGGBB)"})
						return
					}
					set["theme.dark.primaryColor"] = *req.Theme.Dark.PrimaryColor
				}
				if req.Theme.Dark.BackgroundColor != nil {
					if !isValidColorHex(*req.Theme.Dark.BackgroundColor) {
						c.JSON(http.StatusBadRequest, gin.H{"error": "theme.dark.backgroundColor must be a valid hex color (#RRGGBB)"})
						return
					}
					set["theme.dark.backgroundColor"] = *req.Theme.Dark.BackgroundColor
				}
				if req.Theme.Dark.TextColor != nil {
					if !isValidColorHex(*req.Theme.Dark.TextColor) {
						c.JSON(http.StatusBadRequest, gin.H{"error": "theme.dark.textColor must be a valid hex color (#RRGGBB)"})
						return
					}
					set["theme.dark.textColor"] = *req.Theme.Dark.TextColor
				}
				if req.Theme.Dark.AccentColor != nil {
					if !isValidColorHex(*req.Theme.Dark.AccentColor) {
						c.JSON(http.StatusBadRequest, gin.H{"error": "theme.dark.accentColor must be a valid hex color (#RRGGBB)"})
						return
					}
					set["theme.dark.accentColor"] = *req.Theme.Dark.AccentColor
				}
			}
			if req.Theme.Typography != nil {
				if req.Theme.Typography.FontFamily != nil {
					set["theme.typography.fontFamily"] = *req.Theme.Typography.FontFamily
				}
				if req.Theme.Typography.FontScale != nil {
					if *req.Theme.Typography.FontScale != "sm" && *req.Theme.Typography.FontScale != "md" && *req.Theme.Typography.FontScale != "lg" {
						c.JSON(http.StatusBadRequest, gin.H{"error": "theme.typography.fontScale must be one of: sm, md, lg"})
						return
					}
					set["theme.typography.fontScale"] = *req.Theme.Typography.FontScale
				}
			}
		}

		if req.Layout != nil && req.Layout.Variant != nil {
			if *req.Layout.Variant != "classic" && *req.Layout.Variant != "compact" && *req.Layout.Variant != "minimal" && *req.Layout.Variant != "cards" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "layout.variant must be one of: classic, compact, minimal, cards"})
				return
			}
			set["layout.variant"] = *req.Layout.Variant
		}

		if req.Footer != nil {
			if req.Footer.Text != nil {
				set["footer.text"] = *req.Footer.Text
			}
			if req.Footer.ShowPoweredBy != nil {
				set["footer.showPoweredBy"] = *req.Footer.ShowPoweredBy
			}
		}

		if req.SSO != nil {
			if req.SSO.Enabled != nil {
				set["sso.enabled"] = *req.SSO.Enabled
			}
			if req.SSO.Provider != nil {
				set["sso.provider"] = strings.TrimSpace(*req.SSO.Provider)
			}
			if req.SSO.Issuer != nil {
				set["sso.issuer"] = strings.TrimSpace(*req.SSO.Issuer)
			}
			if req.SSO.Audience != nil {
				set["sso.audience"] = strings.TrimSpace(*req.SSO.Audience)
			}
			if req.SSO.Algorithm != nil {
				algorithm := strings.ToUpper(strings.TrimSpace(*req.SSO.Algorithm))
				if algorithm != "HS256" && algorithm != "RS256" {
					c.JSON(http.StatusBadRequest, gin.H{"error": "sso.algorithm must be HS256 or RS256"})
					return
				}
				set["sso.algorithm"] = algorithm
			}
			if req.SSO.SharedSecret != nil {
				set["sso.sharedSecret"] = strings.TrimSpace(*req.SSO.SharedSecret)
			}
			if req.SSO.PublicKeyPEM != nil {
				set["sso.publicKeyPem"] = strings.TrimSpace(*req.SSO.PublicKeyPEM)
			}
		}

		if req.CustomCSS != nil {
			set["customCss"] = *req.CustomCSS
		}

		if len(set) == 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no updatable fields provided"})
			return
		}

		updatedSSO, err := settingsRepo.UpdateSSOSettings(ctx, set)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		updated := current
		updated.UpdatedAt = set["updatedAt"].(time.Time)
		updated.SSO = *updatedSSO

		err = settingsCollection(db).FindOne(ctx, bson.M{"key": models.StatusPageSettingsKey}).Decode(&updated)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		BroadcastEvent(hub, "status_page_settings_updated", updated)
		c.JSON(http.StatusOK, buildAdminStatusPageSettingsResponse(updated))
	}
}
