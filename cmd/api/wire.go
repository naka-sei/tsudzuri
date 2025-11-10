//go:build wireinject

package main

import (
	"github.com/google/wire"

	pagerepo "github.com/naka-sei/tsudzuri/infrastructure/db/page"
	ipostgres "github.com/naka-sei/tsudzuri/infrastructure/db/postgres"
	userrepo "github.com/naka-sei/tsudzuri/infrastructure/db/user"
	httpserver "github.com/naka-sei/tsudzuri/presentation/http"
	pagehandler "github.com/naka-sei/tsudzuri/presentation/http/page"
	userhandler "github.com/naka-sei/tsudzuri/presentation/http/user"
	pageusecase "github.com/naka-sei/tsudzuri/usecase/page"
	useservice "github.com/naka-sei/tsudzuri/usecase/service"
	userusecase "github.com/naka-sei/tsudzuri/usecase/user"
)

func InitializePresentationServer(
	dbConn *ipostgres.Connection,
) (*httpserver.Server, error) {
	wire.Build(
		presentationSet,
		usecaseSet,
		repoSet,
		serviceSet,
	)
	return nil, nil
}

func transactionServiceProvider(dbc *ipostgres.Connection) useservice.TransactionService {
	return dbc
}

var (
	presentationSet = wire.NewSet(
		pagehandler.NewCreateService,
		pagehandler.NewGetService,
		pagehandler.NewListService,
		pagehandler.NewEditService,
		pagehandler.NewDeleteService,
		pagehandler.NewLinkAddService,
		pagehandler.NewLinkRemoveService,
		userhandler.NewCreateService,
		userhandler.NewLoginService,
		httpserver.NewServer,
	)
	usecaseSet = wire.NewSet(
		pageusecase.NewCreateUsecase,
		pageusecase.NewGetUsecase,
		pageusecase.NewListUsecase,
		pageusecase.NewEditUsecase,
		pageusecase.NewDeleteUsecase,
		pageusecase.NewLinkAddUsecase,
		pageusecase.NewLinkRemoveUsecase,
		userusecase.NewCreateUsecase,
		userusecase.NewLoginUsecase,
	)
	repoSet = wire.NewSet(
		pagerepo.NewPageRepository,
		userrepo.NewUserRepository,
	)
	serviceSet = wire.NewSet(
		transactionServiceProvider,
	)
)
