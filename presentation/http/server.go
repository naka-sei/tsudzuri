package http

import (
	duser "github.com/naka-sei/tsudzuri/domain/user"
	"github.com/naka-sei/tsudzuri/pkg/cache"
	page "github.com/naka-sei/tsudzuri/presentation/http/page"
	user "github.com/naka-sei/tsudzuri/presentation/http/user"
	prouter "github.com/naka-sei/tsudzuri/presentation/router"
)

const (
	apiPathPrefix = "/api"
	versionPathV1 = "/v1"
	apiV1BasePath = apiPathPrefix + versionPathV1

	// Page routes
	pageBasePath  = "/pages"
	pageIDPath    = "/{id}"
	pageLinksPath = "/links"

	// User routes
	userBasePath  = "/users"
	userLoginPath = "/login"
)

type Server struct {
	pageCreate     *page.CreateService
	pageGet        *page.GetService
	pageList       *page.ListService
	pageEdit       *page.EditService
	pageDelete     *page.DeleteService
	pageLinkAdd    *page.LinkAddService
	pageLinkRemove *page.LinkRemoveService
	userCreate     *user.CreateService
	userLogin      *user.LoginService
}

func NewServer(
	createPageService *page.CreateService,
	getPageService *page.GetService,
	listPageService *page.ListService,
	editPageService *page.EditService,
	deletePageService *page.DeleteService,
	linkAddPageService *page.LinkAddService,
	linkRemovePageService *page.LinkRemoveService,
	createUserService *user.CreateService,
	loginUserService *user.LoginService,
) *Server {
	return &Server{
		pageCreate:     createPageService,
		pageGet:        getPageService,
		pageList:       listPageService,
		pageEdit:       editPageService,
		pageDelete:     deletePageService,
		pageLinkAdd:    linkAddPageService,
		pageLinkRemove: linkRemovePageService,
		userCreate:     createUserService,
		userLogin:      loginUserService,
	}
}

// WithUserCache configures the shared user cache for authentication-aware components.
func (s *Server) WithUserCache(c cache.Cache[*duser.User]) {
	if s == nil {
		return
	}
	if s.userCreate != nil {
		s.userCreate.SetCache(c)
	}
}

func (s *Server) Route(r prouter.Router) {
	// /api/v1
	r.Route(apiV1BasePath, func(r prouter.Router) {
		// /api/v1/pages
		r.Route(pageBasePath, func(r prouter.Router) {
			// POST /api/v1/pages
			r.Post("/", s.pageCreate.Create, prouter.WithStatusCreated())
			// GET /api/v1/pages
			r.Get("/", s.pageList.List)
			// /api/v1/pages/{id}
			r.Route(pageIDPath, func(r prouter.Router) {
				// GET /api/v1/pages/{id}
				r.Get("/", s.pageGet.Get)
				// PUT /api/v1/pages/{id}
				r.Put("/", s.pageEdit.Edit)
				// DELETE /api/v1/pages/{id}
				r.Delete("/", s.pageDelete.Delete, prouter.WithStatusNoContent())
				// /api/v1/pages/{id}/links
				r.Route(pageLinksPath, func(r prouter.Router) {
					// POST /api/v1/pages/{id}/links
					r.Post("/", s.pageLinkAdd.LinkAdd, prouter.WithStatusCreated())
					// DELETE /api/v1/pages/{id}/links
					r.Delete("/", s.pageLinkRemove.LinkRemove, prouter.WithStatusNoContent())
				})
			})
		})
		// /api/v1/users
		r.Route(userBasePath, func(r prouter.Router) {
			// POST /api/v1/users
			r.Post("/", s.userCreate.Create, prouter.WithStatusCreated())
			// POST /api/v1/users/login
			r.Post(userLoginPath, s.userLogin.Login)
		})
	})
}
