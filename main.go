package main

import (
    "errors"
    "strings"
    "sync"
)

type Player struct {
    name       string
    currentMap *Map
    msgChan    chan string
    mu         sync.RWMutex
}

type Map struct {
    id      int
    players map[string]*Player
    msgChan chan string
    mu      sync.RWMutex
}

type Game struct {
    maps    map[int]*Map
    players map[string]*Player
    mu      sync.RWMutex
}

func NewGame(mapIds []int) (*Game, error) {
    for _, id := range mapIds {
        if id <= 0 {
            return nil, errors.New("map id must be positive")
        }
    }

    game := &Game{
        maps:    make(map[int]*Map),
        players: make(map[string]*Player),
    }

    for _, id := range mapIds {
        m := &Map{
            id:      id,
            players: make(map[string]*Player),
            msgChan: make(chan string, 100),
        }
        game.maps[id] = m
        go m.FanOutMessages()
    }

    return game, nil
}

func (g *Game) ConnectPlayer(name string) error {
    if name == "" {
        return errors.New("player name cannot be empty")
    }

    normalizedName := strings.ToLower(name)
    
    g.mu.Lock()
    if _, exists := g.players[normalizedName]; exists {
        g.mu.Unlock()
        return errors.New("player already exists")
    }

    player := &Player{
        name:    name,
        msgChan: make(chan string, 100),
    }
    g.players[normalizedName] = player
    g.mu.Unlock()
    
    return nil
}

func (g *Game) SwitchPlayerMap(name string, mapId int) error {
    normalizedName := strings.ToLower(name)
    
    g.mu.RLock()
    player, exists := g.players[normalizedName]
    if !exists {
        g.mu.RUnlock()
        return errors.New("player not found")
    }

    newMap, exists := g.maps[mapId]
    if !exists {
        g.mu.RUnlock()
        return errors.New("map not found")
    }
    g.mu.RUnlock()

    player.mu.Lock()
    oldMap := player.currentMap
    if oldMap != nil {
        if oldMap.id == mapId {
            player.mu.Unlock()
            return errors.New("player already in this map")
        }
    }

    // First update the player's map reference
    player.currentMap = newMap
    player.mu.Unlock()

    // Then handle the map players' lists
    if oldMap != nil {
        oldMap.mu.Lock()
        delete(oldMap.players, normalizedName)
        oldMap.mu.Unlock()
    }

    newMap.mu.Lock()
    newMap.players[normalizedName] = player
    newMap.mu.Unlock()

    return nil
}

func (g *Game) GetPlayer(name string) (*Player, error) {
    g.mu.RLock()
    defer g.mu.RUnlock()

    if player, exists := g.players[strings.ToLower(name)]; exists {
        return player, nil
    }
    return nil, errors.New("player not found")
}

func (g *Game) GetMap(mapId int) (*Map, error) {
    g.mu.RLock()
    defer g.mu.RUnlock()

    if m, exists := g.maps[mapId]; exists {
        return m, nil
    }
    return nil, errors.New("map not found")
}

func (m *Map) FanOutMessages() {
    for msg := range m.msgChan {
        m.mu.RLock()
        currentPlayers := make(map[string]*Player)
        for name, player := range m.players {
            currentPlayers[name] = player
        }
        m.mu.RUnlock()

        parts := strings.SplitN(msg, " says: ", 2)
        senderName := strings.ToLower(strings.TrimSpace(parts[0]))

        for name, player := range currentPlayers {
            if name != senderName {
                select {
                case player.msgChan <- msg:
                default:
                    // Skip if channel is full
                }
            }
        }
    }
}

func (p *Player) GetChannel() <-chan string {
    return p.msgChan
}

func (p *Player) SendMessage(msg string) error {
    p.mu.RLock()
    currentMap := p.currentMap
    p.mu.RUnlock()

    if currentMap == nil {
        return errors.New("player not in any map")
    }

    formattedName := strings.Title(strings.ToLower(p.name))
    formattedMsg := formattedName + " says: " + msg

    select {
    case currentMap.msgChan <- formattedMsg:
        return nil
    default:
        return errors.New("message channel is full")
    }
}

func (p *Player) GetName() string {
    return p.name
}