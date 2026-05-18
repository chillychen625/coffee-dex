package game

import "go-coffee-log/service"

// Services bundles all service-layer references the game screens need.
type Services struct {
	Coffee     *service.CoffeeService
	Brew       *service.BrewService
	Brewer     *service.BrewerService
	Pokemon    *service.PokemonService
	Statistics *service.StatisticsService
}
