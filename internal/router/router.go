package router

import (
	"github.com/go-redis/redis/v8"
	kxlcfg "github.com/linkyfish/kxl_backend_go/internal/config"
	"github.com/linkyfish/kxl_backend_go/internal/handler"
	admin "github.com/linkyfish/kxl_backend_go/internal/handler/api/admin"
	v1 "github.com/linkyfish/kxl_backend_go/internal/handler/api/v1"
	"github.com/linkyfish/kxl_backend_go/internal/handler/upload"
	kxlweb "github.com/linkyfish/kxl_backend_go/internal/handler/web"
	kxlmw "github.com/linkyfish/kxl_backend_go/internal/middleware"
	"github.com/linkyfish/kxl_backend_go/internal/service"
	kxlvalidator "github.com/linkyfish/kxl_backend_go/internal/validator"
	"github.com/linkyfish/kxl_backend_go/pkg/session"
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	"gorm.io/gorm"
)

type Deps struct {
	Cfg    *kxlcfg.Config
	DB     *gorm.DB
	Redis  *redis.Client
	Sess   *session.Manager
}

// New creates an Echo instance with all API routes registered.
func New(deps Deps) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.Validator = kxlvalidator.New()

	// SSR templates (Pongo2).
	renderer, err := kxlweb.NewRenderer("templates")
	if err != nil {
		e.Logger.Fatalf("init templates renderer: %v", err)
	}
	e.Renderer = renderer

	e.HTTPErrorHandler = kxlmw.NewHTTPErrorHandler(deps.Cfg)

	e.Use(echomw.Recover())
	e.Use(echomw.Logger())
	e.Use(kxlmw.CORS(deps.Cfg))
	e.Use(kxlmw.RateLimit(deps.Redis, deps.Cfg))

	// Services (thin wrappers around GORM).
	authSvc := service.NewAuthService(deps.DB)
	userSvc := service.NewUserService(deps.DB)
	articleSvc := service.NewArticleService(deps.DB)
	projectSvc := service.NewProjectService(deps.DB)
	caseSvc := service.NewCaseService(deps.DB)
	messageSvc := service.NewMessageService(deps.DB)
	settingsSvc := service.NewSettingsService(deps.DB)
	rbacSvc := service.NewRbacService(deps.DB, deps.Redis, deps.Cfg.Security.RbacCacheTTLSeconds)
	uploadSvc := service.NewUploadService(deps.Cfg)
	bannerSvc := service.NewBannerService(deps.DB)
	testimonialSvc := service.NewTestimonialService(deps.DB)
	solutionSvc := service.NewSolutionService(deps.DB)
	partnerSvc := service.NewPartnerService(deps.DB)
	friendlySvc := service.NewFriendlyLinkService(deps.DB)
	searchSvc := service.NewSearchService(deps.DB)
	systemConfigSvc := service.NewSystemConfigService(deps.DB)

	// Health checks.
	health := &handler.HealthHandler{DB: deps.DB, Redis: deps.Redis}
	e.GET("/health", health.Health)
	e.GET("/ready", health.Ready)

	// Static assets and uploads.
	e.Static("/static", "static")
	if deps.Cfg != nil && deps.Cfg.Uploads.Dir != "" {
		e.Static("/uploads", deps.Cfg.Uploads.Dir)
	} else {
		e.Static("/uploads", "uploads")
	}

	// SSR website routes.
	home := &kxlweb.HomeHandler{
		Settings:     settingsSvc,
		Banners:      bannerSvc,
		Projects:     projectSvc,
		Cases:        caseSvc,
		Articles:     articleSvc,
		Testimonials: testimonialSvc,
		Solutions:    solutionSvc,
		Partners:     partnerSvc,
		Friendly:     friendlySvc,
	}
	e.GET("/", home.Index)

	about := &kxlweb.AboutHandler{Settings: settingsSvc, Friendly: friendlySvc}
	e.GET("/about", about.Index)

	webArticles := &kxlweb.ArticleHandler{DB: deps.DB, Settings: settingsSvc, Friendly: friendlySvc, Articles: articleSvc}
	e.GET("/articles", webArticles.List)
	e.GET("/articles/:id", webArticles.Detail)

	webProjects := &kxlweb.ProjectHandler{DB: deps.DB, Settings: settingsSvc, Friendly: friendlySvc, Projects: projectSvc}
	e.GET("/projects", webProjects.List)
	e.GET("/projects/:id", webProjects.Detail)

	webCases := &kxlweb.CaseHandler{DB: deps.DB, Settings: settingsSvc, Friendly: friendlySvc, Cases: caseSvc, Projects: projectSvc}
	e.GET("/cases", webCases.List)
	e.GET("/cases/:id", webCases.Detail)

	contact := &kxlweb.ContactHandler{Settings: settingsSvc, Friendly: friendlySvc, Messages: messageSvc}
	e.GET("/contact", contact.Index)
	e.POST("/contact/submit", contact.Submit)

	webSearch := &kxlweb.SearchHandler{Settings: settingsSvc, Friendly: friendlySvc, Search: searchSvc}
	e.GET("/search", webSearch.Index)

	webAuth := &kxlweb.AuthHandler{Settings: settingsSvc, Friendly: friendlySvc, Auth: authSvc, Sessions: deps.Sess}
	e.GET("/login", webAuth.LoginPage)
	e.POST("/login", webAuth.LoginSubmit)
	e.GET("/register", webAuth.RegisterPage)
	e.POST("/register", webAuth.RegisterSubmit)
	e.GET("/logout", webAuth.Logout)

	// Public API (/api/v1/*).
	v1Group := e.Group("/api/v1")
	{
		// Auth endpoints.
		authHandler := &v1.AuthHandler{Auth: authSvc, Sessions: deps.Sess}
		v1Group.POST("/auth/register", authHandler.Register)
		v1Group.POST("/auth/login", authHandler.Login)

		userAuthed := v1Group.Group("", kxlmw.AuthUser(deps.DB, deps.Sess))
		userAuthed.POST("/auth/logout", authHandler.Logout)
		userAuthed.GET("/auth/me", authHandler.Me)

		// User endpoints.
		userHandler := &v1.UserHandler{DB: deps.DB, Sessions: deps.Sess}
		userAuthed.POST("/users/change-password", userHandler.ChangePassword)

		// Content endpoints.
		articleHandler := &v1.ArticleHandler{DB: deps.DB, Articles: articleSvc}
		v1Group.GET("/articles", articleHandler.List)
		v1Group.GET("/articles/:id", articleHandler.Detail)
		v1Group.GET("/articles/:id/related", articleHandler.Related)
		v1Group.GET("/articles/:id/navigation", articleHandler.Navigation)

		projectHandler := &v1.ProjectHandler{DB: deps.DB, Projects: projectSvc}
		v1Group.GET("/projects", projectHandler.List)
		v1Group.GET("/projects/:id", projectHandler.Detail)

		caseHandler := &v1.CaseHandler{DB: deps.DB, Cases: caseSvc, Projects: projectSvc}
		v1Group.GET("/cases", caseHandler.List)
		v1Group.GET("/cases/:id", caseHandler.Detail)

		messageHandler := &v1.MessageHandler{Messages: messageSvc}
		v1Group.POST("/messages", messageHandler.Submit)

		settingsHandler := &v1.SettingsHandler{Settings: settingsSvc}
		v1Group.GET("/company-info", settingsHandler.GetCompanyInfo)
		v1Group.GET("/milestones", settingsHandler.ListMilestones)
		v1Group.GET("/team-members", settingsHandler.ListTeam)

		bannerHandler := &v1.BannerHandler{Banners: bannerSvc}
		v1Group.GET("/banners", bannerHandler.List)

		testimonialHandler := &v1.TestimonialHandler{Testimonials: testimonialSvc}
		v1Group.GET("/testimonials", testimonialHandler.List)

		solutionHandler := &v1.SolutionHandler{Solutions: solutionSvc}
		v1Group.GET("/solutions", solutionHandler.List)

		partnerHandler := &v1.PartnerHandler{Partners: partnerSvc}
		v1Group.GET("/partners", partnerHandler.List)

		friendlyHandler := &v1.FriendlyLinkHandler{FriendlyLinks: friendlySvc}
		v1Group.GET("/friendly-links", friendlyHandler.List)

		searchHandler := &v1.SearchHandler{SearchSvc: searchSvc}
		v1Group.GET("/search", searchHandler.Search)
		v1Group.GET("/search/suggestions", searchHandler.Suggestions)
	}

	// Upload API (shared path prefix like Rust/PHP backends).
	uploadHandler := &upload.UploadHandler{Uploads: uploadSvc}
	e.POST("/api/upload/image", uploadHandler.UploadImage)
	e.POST("/api/upload/video", uploadHandler.UploadVideo)

	// Admin API (/api/admin/*).
	adminGroup := e.Group("/api/admin")
	{
		adminAuthHandler := &admin.AuthHandler{Auth: authSvc, RBAC: rbacSvc, Sessions: deps.Sess}
		adminGroup.POST("/auth/login", adminAuthHandler.Login)

		adminAuthed := adminGroup.Group("", kxlmw.AuthAdmin(deps.DB, deps.Sess, rbacSvc))
		adminAuthed.POST("/auth/logout", adminAuthHandler.Logout)
		adminAuthed.GET("/auth/me", adminAuthHandler.Me)

		userHandler := &admin.UserHandler{Users: userSvc}
		adminAuthed.GET("/users", userHandler.ListUsers)
		adminAuthed.GET("/users/:id", userHandler.DetailUser)
		adminAuthed.PATCH("/users/:id/status", userHandler.UpdateUserStatus)

		adminAuthed.GET("/admins", userHandler.ListAdmins)
		adminAuthed.POST("/admins", userHandler.CreateAdmin)
		adminAuthed.PUT("/admins/:id", userHandler.UpdateAdmin)
		adminAuthed.DELETE("/admins/:id", userHandler.DeleteAdmin)

		articleAdminHandler := &admin.ArticleHandler{DB: deps.DB, Articles: articleSvc}
		adminAuthed.GET("/articles", articleAdminHandler.List)
		adminAuthed.GET("/articles/:id", articleAdminHandler.Detail)
		adminAuthed.POST("/articles", articleAdminHandler.Create)
		adminAuthed.PUT("/articles/:id", articleAdminHandler.Update)
		adminAuthed.DELETE("/articles/:id", articleAdminHandler.Delete)
		adminAuthed.PATCH("/articles/:id/status", articleAdminHandler.UpdateStatus)
		adminAuthed.PUT("/articles/:id/tags", articleAdminHandler.SetTags)

		projectHandler := &admin.ProjectHandler{DB: deps.DB, Projects: projectSvc}
		adminAuthed.GET("/projects", projectHandler.List)
		adminAuthed.POST("/projects", projectHandler.Create)
		adminAuthed.PUT("/projects/:id", projectHandler.Update)
		adminAuthed.DELETE("/projects/:id", projectHandler.Delete)
		adminAuthed.PATCH("/projects/:id/status", projectHandler.UpdateStatus)
		adminAuthed.GET("/projects/:id/features", projectHandler.ListFeatures)
		adminAuthed.POST("/projects/:id/features", projectHandler.CreateFeature)
		adminAuthed.PUT("/projects/:id/features/:feature_id", projectHandler.UpdateFeature)
		adminAuthed.DELETE("/projects/:id/features/:feature_id", projectHandler.DeleteFeature)
		adminAuthed.GET("/projects/:id/media", projectHandler.ListMedia)
		adminAuthed.POST("/projects/:id/media", projectHandler.CreateMedia)
		adminAuthed.PUT("/projects/:id/media/:media_id", projectHandler.UpdateMedia)
		adminAuthed.DELETE("/projects/:id/media/:media_id", projectHandler.DeleteMedia)
		adminAuthed.GET("/projects/:id/versions", projectHandler.ListVersions)
		adminAuthed.POST("/projects/:id/versions", projectHandler.CreateVersion)
		adminAuthed.PUT("/projects/:id/versions/:version_id", projectHandler.UpdateVersion)
		adminAuthed.DELETE("/projects/:id/versions/:version_id", projectHandler.DeleteVersion)
		adminAuthed.PUT("/projects/:id/tags", projectHandler.SetTags)

		caseAdminHandler := &admin.CaseHandler{DB: deps.DB, Cases: caseSvc, Projects: projectSvc}
		adminAuthed.GET("/cases", caseAdminHandler.List)
		adminAuthed.GET("/cases/:id", caseAdminHandler.Detail)
		adminAuthed.POST("/cases", caseAdminHandler.Create)
		adminAuthed.PUT("/cases/:id", caseAdminHandler.Update)
		adminAuthed.DELETE("/cases/:id", caseAdminHandler.Delete)
		adminAuthed.PATCH("/cases/:id/status", caseAdminHandler.UpdateStatus)
		adminAuthed.PUT("/cases/:id/projects", caseAdminHandler.SetProjects)

		messageHandler := &admin.MessageHandler{Messages: messageSvc}
		adminAuthed.GET("/messages", messageHandler.List)
		adminAuthed.GET("/messages/:id", messageHandler.Detail)
		adminAuthed.PATCH("/messages/:id/status", messageHandler.UpdateStatus)
		adminAuthed.PATCH("/messages/:id/note", messageHandler.UpdateNote)
		adminAuthed.DELETE("/messages/:id", messageHandler.Delete)
		adminAuthed.DELETE("/messages", messageHandler.BatchDelete)
		adminAuthed.GET("/messages/stats", messageHandler.Stats)

		settingsHandler := &admin.SettingsHandler{Settings: settingsSvc}
		adminAuthed.GET("/categories", settingsHandler.ListCategories)
		adminAuthed.POST("/categories", settingsHandler.CreateCategory)
		adminAuthed.PUT("/categories/:id", settingsHandler.UpdateCategory)
		adminAuthed.DELETE("/categories/:id", settingsHandler.DeleteCategory)
		adminAuthed.GET("/tags", settingsHandler.ListTags)
		adminAuthed.POST("/tags", settingsHandler.CreateTag)
		adminAuthed.PUT("/tags/:id", settingsHandler.UpdateTag)
		adminAuthed.DELETE("/tags/:id", settingsHandler.DeleteTag)
		adminAuthed.PUT("/company-info", settingsHandler.UpdateCompanyInfo)
		adminAuthed.GET("/milestones", settingsHandler.ListMilestones)
		adminAuthed.POST("/milestones", settingsHandler.CreateMilestone)
		adminAuthed.PUT("/milestones/:id", settingsHandler.UpdateMilestone)
		adminAuthed.DELETE("/milestones/:id", settingsHandler.DeleteMilestone)
		adminAuthed.GET("/team-members", settingsHandler.ListTeam)
		adminAuthed.POST("/team-members", settingsHandler.CreateTeamMember)
		adminAuthed.PUT("/team-members/:id", settingsHandler.UpdateTeamMember)
		adminAuthed.DELETE("/team-members/:id", settingsHandler.DeleteTeamMember)
		adminAuthed.GET("/dashboard/stats", settingsHandler.DashboardStats)

		rbacHandler := &admin.RbacHandler{RBAC: rbacSvc}
		adminAuthed.GET("/rbac/roles", rbacHandler.ListRoles)
		adminAuthed.POST("/rbac/roles", rbacHandler.CreateRole)
		adminAuthed.PUT("/rbac/roles/:code", rbacHandler.UpdateRole)
		adminAuthed.DELETE("/rbac/roles/:code", rbacHandler.DeleteRole)
		adminAuthed.GET("/rbac/permissions", rbacHandler.ListPermissions)
		adminAuthed.GET("/rbac/roles/:code/permissions", rbacHandler.ListRolePermissions)
		adminAuthed.PUT("/rbac/roles/:code/permissions", rbacHandler.SetRolePermissions)

		adminUploadHandler := &admin.UploadHandler{Uploads: uploadSvc}
		adminAuthed.POST("/upload/image", adminUploadHandler.UploadImage)
		adminAuthed.POST("/upload/video", adminUploadHandler.UploadVideo)
		adminAuthed.DELETE("/upload/*", adminUploadHandler.DeleteFile)

		bannerAdminHandler := &admin.BannerHandler{Banners: bannerSvc}
		adminAuthed.GET("/banners", bannerAdminHandler.List)
		adminAuthed.GET("/banners/:id", bannerAdminHandler.Detail)
		adminAuthed.POST("/banners", bannerAdminHandler.Create)
		adminAuthed.PUT("/banners/:id", bannerAdminHandler.Update)
		adminAuthed.DELETE("/banners/:id", bannerAdminHandler.Delete)

		testimonialAdminHandler := &admin.TestimonialHandler{Testimonials: testimonialSvc}
		adminAuthed.GET("/testimonials", testimonialAdminHandler.List)
		adminAuthed.GET("/testimonials/:id", testimonialAdminHandler.Detail)
		adminAuthed.POST("/testimonials", testimonialAdminHandler.Create)
		adminAuthed.PUT("/testimonials/:id", testimonialAdminHandler.Update)
		adminAuthed.DELETE("/testimonials/:id", testimonialAdminHandler.Delete)

		solutionAdminHandler := &admin.SolutionHandler{Solutions: solutionSvc}
		adminAuthed.GET("/solutions", solutionAdminHandler.List)
		adminAuthed.GET("/solutions/:id", solutionAdminHandler.Detail)
		adminAuthed.POST("/solutions", solutionAdminHandler.Create)
		adminAuthed.PUT("/solutions/:id", solutionAdminHandler.Update)
		adminAuthed.DELETE("/solutions/:id", solutionAdminHandler.Delete)

		partnerAdminHandler := &admin.PartnerHandler{Partners: partnerSvc}
		adminAuthed.GET("/partners", partnerAdminHandler.List)
		adminAuthed.GET("/partners/:id", partnerAdminHandler.Detail)
		adminAuthed.POST("/partners", partnerAdminHandler.Create)
		adminAuthed.PUT("/partners/:id", partnerAdminHandler.Update)
		adminAuthed.DELETE("/partners/:id", partnerAdminHandler.Delete)

		friendlyAdminHandler := &admin.FriendlyLinkHandler{FriendlyLinks: friendlySvc}
		adminAuthed.GET("/friendly-links", friendlyAdminHandler.List)
		adminAuthed.GET("/friendly-links/:id", friendlyAdminHandler.Detail)
		adminAuthed.POST("/friendly-links", friendlyAdminHandler.Create)
		adminAuthed.PUT("/friendly-links/:id", friendlyAdminHandler.Update)
		adminAuthed.DELETE("/friendly-links/:id", friendlyAdminHandler.Delete)
		adminAuthed.DELETE("/friendly-links", friendlyAdminHandler.BatchDelete)

		systemConfigHandler := &admin.SystemConfigHandler{SystemConfigs: systemConfigSvc}
		adminAuthed.GET("/system-configs", systemConfigHandler.List)
		adminAuthed.POST("/system-configs", systemConfigHandler.Create)
		adminAuthed.PUT("/system-configs/:id", systemConfigHandler.Update)
		adminAuthed.DELETE("/system-configs/:id", systemConfigHandler.Delete)
	}

	return e
}
