# cache-for-discord
A better cache for discordgo, memory friendly and good for log systems, this file is similar to the [discordgo](https://github.com/bwmarrin/discordgo) state.go file

## Installation
Download the state.go file or use these commands:
```go
go get github.com/meszmate/cache-for-discord
go get github.com/bwmarrin/discordgo
```

## Example

```go
package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/meszmate/cache-for-discord"
)
var shardSessions = make([]*discordgo.Session, 0)
var cache *dcache.State = dcache.NewState()
const ShardCount int = 3
func main() {
	for i := 0; i < shardCount; i++ {
		cache.CreateNewShard(i)
		dg, err := discordgo.New("Bot BOT_TOKEN")
		if err != nil {
			fmt.Printf("error when starting shard %d: %s\n", i, err)
			continue
		}
		dg.ShardID = i
		dg.ShardCount = shardCount
		dg.Identify.Intents = discordgo.IntentGuilds | discordgo.IntentsGuilds | discordgo.IntentGuildMessages | discordgo.IntentsGuildMessages | discordgo.IntentsGuildMembers | discordgo.IntentGuildMembers | discordgo.IntentGuildModeration // etc...
		dg.StateEnabled = false

		dg.AddHandler(MemberChunkHandler)
		dg.AddHandler(loadGuilds)
		dg.AddHandler(wsHandleMemberAdd)
		dg.AddHandler(wsHandleMemberRemove)
		dg.AddHandler(wsHandleGuildUpdate)
		dg.AddHandler(wsHandleRoleUpdate)
		dg.AddHandler(wsHandleRoleCreate)
		dg.AddHandler(wsHandleRoleRemove)
		dg.AddHandler(wsHandleChannelUpdate)
		dg.AddHandler(wsHandleChannelCreate)
		dg.AddHandler(wsHandleChannelRemove)
		dg.AddHandler(wsHandleMessageAdd)
		dg.AddHandler(wsHandleMessageDelete)
		dg.AddHandler(wsHandleMessageDeleteBulk)
		dg.AddHandler(wsHandleEmojisUpdate)

		err = dg.Open()
		if err != nil {
			fmt.Printf("error when starting shard %d: %s\n", i, err)
			continue
		}
		shardSessions = append(shardSessions, dg)

		fmt.Printf("Shard %d started\n", i)
	}
  fmt.Println("Bot started")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
	for _, session := range shardSessions {
		session.Close()
	}
}

// discord is using lazy loading for guilds, so add to the state if the guild is available or you can use the ready event
func loadGuilds(s *discordgo.Session, m *discordgo.GuildCreate) {
	cache.Shards[s.ShardID].GuildAdd(m.Guild)
	err := s.RequestGuildMembers(m.Guild.ID, "", 0, "", false)
	if err != nil{
		fmt.Println(err.Error())
	}
}

// Add server members and etc
func MemberChunkHandler(s *discordgo.Session, m *discordgo.GuildMembersChunk) {
    for _, member := range m.Members {
		    member.GuildID = m.GuildID
        err := cache.Shards[s.ShardID].MemberAdd(member)
        if err != nil {
            fmt.Printf("Failed to add member %s (%s): %s\n", member.User.Username, member.User.ID, err)
        }
        err := cache.UserAdd(member.User)
        if err != nil {
            fmt.Printf("Failed to add user %s (%s): %s\n", member.User.Username, member.User.ID, err)
        }
    }
}

func wsHandleMemberAdd(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
	g, _ := cache.Shards[s.ShardID].Guild(m.GuildID)
	g.MemberCount++
	cache.Shards[s.ShardID].MemberAdd(m.Member)
}
func wsHandleMemberRemove(s *discordgo.Session, m *discordgo.GuildMemberRemove) {
	g, _ := cache.Shards[s.ShardID].Guild(m.GuildID)
	g.MemberCount--
	cache.Shards[s.ShardID].MemberRemove(m.Member)
}
func wsHandleGuildUpdate(s *discordgo.Session, m *discordgo.GuildUpdate) {
	cache.Shards[s.ShardID].GuildAdd(m.Guild)
}
func wsHandleRoleUpdate(s *discordgo.Session, m *discordgo.GuildRoleUpdate) {
	cache.Shards[s.ShardID].RoleAdd(m.GuildID, m.Role)
}
func wsHandleRoleCreate(s *discordgo.Session, m *discordgo.GuildRoleCreate) {
	cache.Shards[s.ShardID].RoleAdd(m.GuildID, m.Role)
}
func wsHandleRoleRemove(s *discordgo.Session, m *discordgo.GuildRoleDelete) {
	cache.Shards[s.ShardID].RoleRemove(m.GuildID, m.RoleID)
}
func wsHandleChannelUpdate(s *discordgo.Session, m *discordgo.ChannelUpdate) {
	cache.Shards[s.ShardID].ChannelAdd(m.GuildID, m.Channel)
}
func wsHandleChannelCreate(s *discordgo.Session, m *discordgo.ChannelCreate) {
	cache.Shards[s.ShardID].ChannelAdd(m.GuildID, m.Channel)
}
func wsHandleChannelRemove(s *discordgo.Session, m *discordgo.ChannelDelete) {
	cache.Shards[s.ShardID].ChannelRemove(m.GuildID, m.Channel.ID)
}
func wsHandleMessageAdd(s *discordgo.Session, m *discordgo.MessageCreate) {
	cache.Shards[s.ShardID].MessageAdd(m.Message)
}
func wsHandleMessageDelete(s *discordgo.Session, m *discordgo.MessageDelete) {
	cache.Shards[s.ShardID].MessageRemove(m.GuildID, m.Message)
}
func wsHandleMessageDeleteBulk(s *discordgo.Session, m *discordgo.MessageDeleteBulk) {
        for _, i := range m.Messages{
	        cache.Shards[s.ShardID].MessageRemove(m.GuildID, i)
        }
}
func wsHandleEmojisUpdate(s *discordgo.Session, m *discordgo.GuildEmojisUpdate){
        cache.Shards[s.ShardID].EmojisUpdate(m.GuildID, m.Emojis)
}

// etc...

```

If you want to get the before and after value:
```go
func wsHandleGuildUpdate(s *discordgo.Session, m *discordgo.GuildUpdate) {
  // Get the before guild value
  guildbefore, err := cache.Shards[s.ShardID].Guild(m.ID)
  if err != nil{
    fmt.Println(err.Error())
  }
  // after update in the cache
  cache.Shards[s.ShardID].GuildAdd(m.Guild)
}
```

The number of the servers and users example

```go
usersnum := len(cache.Users)
serversnum := 0
for _, sh := range cache.Shards{
  serversnum += len(sh.Guilds)
}
```
