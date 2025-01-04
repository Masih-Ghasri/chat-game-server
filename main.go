package main

type Player struct{}

type Map struct{}

type Game struct{}

func NewGame(mapIds []int) (*Game, error) {
	return nil, nil
}

func (g *Game) ConnectPlayer(name string) error {
	return nil
}

func (g *Game) SwitchPlayerMap(name string, mapId int) error {
	return nil
}

func (g *Game) GetPlayer(name string) (*Player, error) {
	return nil, nil
}

func (g *Game) GetMap(mapId int) (*Map, error) {
	return nil, nil
}

func (m *Map) FanOutMessages() {
}

func (p *Player) GetChannel() <-chan string {
	return nil
}

func (p *Player) SendMessage(msg string) error {
	return nil
}

func (p *Player) GetName() string {
	return ""
}
