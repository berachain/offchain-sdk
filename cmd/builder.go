package cmd

// This package is going to help us build cobra commands

func NewCmd() {

	nba := 5 // new base app()

	nba.RegisterSubber(Job{}, 12)
	nba.RegisterPoller(Job2{}, 12)
	nba.RegisterSubber(Job{}, "0x12345", "LiquidityChanged(string,address,string)")

}
