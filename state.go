package cache

import (
	"errors"
	"sync"

	"github.com/bwmarrin/discordgo"
)

var ErrNilState = errors.New("state not instantiated, please use discordgo.New() or assign Session.State")


var ErrStateNotFound = errors.New("state cache not found")


var ErrMessageIncompletePermissions = errors.New("message incomplete, unable to determine permissions")

type StateData struct {
	sync.RWMutex
	MaxMessageCount    int
	guildMap   map[string]*discordgo.Guild
	memberMap  map[string]map[string]*discordgo.Member
}
type State struct {
	sync.RWMutex
	Users map[string]*discordgo.User
	Shards map[int]*StateData
}

func NewState() *State {
	return &State{
		Users: make(map[string]*discordgo.User),
		Shards: make(map[int]*StateData),
	}
}

func (s *StateData) GuildAdd(guild *discordgo.Guild) error {
	if s == nil {
		return ErrNilState
	}

	s.Lock()
	defer s.Unlock()

	if _, ok := s.memberMap[guild.ID]; !ok {
		members := make(map[string]*discordgo.Member)
		s.memberMap[guild.ID] = members
		
	}
	guild.Members = nil
	if g, ok := s.guildMap[guild.ID]; ok {
		*g = *guild
		return nil
	}

	s.guildMap[guild.ID] = guild

	return nil
}

func (s *StateData) GuildRemove(guild *discordgo.Guild) error {
	if s == nil {
		return ErrNilState
	}

	_, err := s.Guild(guild.ID)
	if err != nil {
		return err
	}

	s.Lock()
	defer s.Unlock()

	delete(s.guildMap, guild.ID)

	return nil
}

func (s *StateData) Guild(guildID string) (*discordgo.Guild, error) {
	if s == nil {
		return nil, ErrNilState
	}

	s.RLock()
	defer s.RUnlock()

	if g, ok := s.guildMap[guildID]; ok {
		return g, nil
	}

	return nil, ErrStateNotFound
}

func (s *State) UserAdd(user *discordgo.User) error {
	s.Lock()
	defer s.Unlock()

	m, ok := s.Users[user.ID]
	if !ok {
		s.Users[user.ID] = user
	} else {
		*m = *user
	}
	return nil
}
func (s *StateData) memberAdd(member *discordgo.Member) error {
	members, ok := s.memberMap[member.GuildID]
	if !ok {
		return ErrStateNotFound
	}

	m, ok := members[member.User.ID]
	if !ok {
		members[member.User.ID] = member
	} else {
		if member.JoinedAt.IsZero() {
			member.JoinedAt = m.JoinedAt
		}
		*m = *member
	}
	return nil
}

func (s *StateData) MemberAdd(member *discordgo.Member) error {
	if s == nil {
		return ErrNilState
	}

	s.Lock()
	defer s.Unlock()

	return s.memberAdd(member)
}

func (s *StateData) MemberRemove(member *discordgo.Member) error {
	if s == nil {
		return ErrNilState
	}

	guild, err := s.Guild(member.GuildID)
	if err != nil {
		return err
	}

	s.Lock()
	defer s.Unlock()

	members, ok := s.memberMap[member.GuildID]
	if !ok {
		return ErrStateNotFound
	}

	_, ok = members[member.User.ID]
	if !ok {
		return ErrStateNotFound
	}
	delete(members, member.User.ID)

	for i, m := range guild.Members {
		if m.User.ID == member.User.ID {
			guild.Members = append(guild.Members[:i], guild.Members[i+1:]...)
			return nil
		}
	}

	return ErrStateNotFound
}

func (s *State) User(userID string) (*discordgo.User, error) {
	if s == nil {
		return nil, ErrNilState
	}

	s.RLock()
	defer s.RUnlock()

	m, ok := s.Users[userID]
	if ok {
		return m, nil
	}

	return nil, ErrStateNotFound
}
func (s *StateData) Member(guildID string, userID string) (*discordgo.Member, error) {
	if s == nil {
		return nil, ErrNilState
	}

	s.RLock()
	defer s.RUnlock()

	members, ok := s.memberMap[guildID]
	if !ok {
		return nil, ErrStateNotFound
	}

	m, ok := members[userID]
	if ok {
		return m, nil
	}

	return nil, ErrStateNotFound
}

func (s *StateData) RoleAdd(guildID string, role *discordgo.Role) error {
	if s == nil {
		return ErrNilState
	}

	guild, err := s.Guild(guildID)
	if err != nil {
		return err
	}

	s.Lock()
	defer s.Unlock()

	for i, r := range guild.Roles {
		if r.ID == role.ID {
			guild.Roles[i] = role
			return nil
		}
	}

	guild.Roles = append(guild.Roles, role)
	return nil
}

func (s *StateData) RoleRemove(guildID string, roleID string) error {
	if s == nil {
		return ErrNilState
	}

	guild, err := s.Guild(guildID)
	if err != nil {
		return err
	}

	s.Lock()
	defer s.Unlock()

	for i, r := range guild.Roles {
		if r.ID == roleID {
			guild.Roles = append(guild.Roles[:i], guild.Roles[i+1:]...)
			return nil
		}
	}

	return ErrStateNotFound
}

func (s *StateData) Role(guildID string, roleID string) (*discordgo.Role, error) {
	if s == nil {
		return nil, ErrNilState
	}

	guild, err := s.Guild(guildID)
	if err != nil {
		return nil, err
	}

	s.RLock()
	defer s.RUnlock()

	for _, r := range guild.Roles {
		if r.ID == roleID {
			return r, nil
		}
	}

	return nil, ErrStateNotFound
}

func (s *StateData) ChannelAdd(guildID string, channel *discordgo.Channel) error {
	if s == nil {
		return ErrNilState
	}

	guild, err := s.Guild(guildID)
	if err != nil {
		return err
	}

	s.Lock()
	defer s.Unlock()

	for i, r := range guild.Channels {
		if r.ID == channel.ID {
			guild.Channels[i] = channel
			return nil
		}
	}

	guild.Channels = append(guild.Channels, channel)
	return nil
}


func (s *StateData) ChannelRemove(guildID string, channelID string) error {
	if s == nil {
		return ErrNilState
	}

	guild, err := s.Guild(guildID)
	if err != nil {
		return err
	}

	s.Lock()
	defer s.Unlock()

	for i, c := range guild.Channels {
		if c.ID == channelID {
			guild.Channels = append(guild.Channels[:i], guild.Channels[i+1:]...)
			return nil
		}
	}

	return ErrStateNotFound
}


func (s *StateData) Channel(guildid string, channelID string) (*discordgo.Channel, error) {
	if s == nil {
		return nil, ErrNilState
	}
	s.RLock()
	defer s.RUnlock()

	guild, err := s.Guild(guildid)
	if err != nil{
		return nil, err
	}
	for _, r := range guild.Channels{
		if r.ID == channelID{
			return r, nil
		}
	}

	return nil, ErrStateNotFound
}


func (s *StateData) Emoji(guildID, emojiID string) (*discordgo.Emoji, error) {
	if s == nil {
		return nil, ErrNilState
	}

	guild, err := s.Guild(guildID)
	if err != nil {
		return nil, err
	}

	s.RLock()
	defer s.RUnlock()

	for _, e := range guild.Emojis {
		if e.ID == emojiID {
			return e, nil
		}
	}

	return nil, ErrStateNotFound
}


func (s *StateData) EmojiAdd(guildID string, emoji *discordgo.Emoji) error {
	if s == nil {
		return ErrNilState
	}

	guild, err := s.Guild(guildID)
	if err != nil {
		return err
	}

	s.Lock()
	defer s.Unlock()

	for i, e := range guild.Emojis {
		if e.ID == emoji.ID {
			guild.Emojis[i] = emoji
			return nil
		}
	}

	guild.Emojis = append(guild.Emojis, emoji)
	return nil
}


func (s *StateData) MessageAdd(message *discordgo.Message) error {
	if s == nil {
		return ErrNilState
	}

	c, err := s.Channel(message.GuildID, message.ChannelID)
	if err != nil {
		return err
	}

	s.Lock()
	defer s.Unlock()

	// If the message exists, merge in the new message contents.
	for _, m := range c.Messages {
		if m.ID == message.ID {
			if message.Content != "" {
				m.Content = message.Content
			}
			if message.EditedTimestamp != nil {
				m.EditedTimestamp = message.EditedTimestamp
			}
			if message.Mentions != nil {
				m.Mentions = message.Mentions
			}
			if message.Embeds != nil {
				m.Embeds = message.Embeds
			}
			if message.Attachments != nil {
				m.Attachments = message.Attachments
			}
			if !message.Timestamp.IsZero() {
				m.Timestamp = message.Timestamp
			}
			if message.Author != nil {
				m.Author = message.Author
			}
			if message.Components != nil {
				m.Components = message.Components
			}

			return nil
		}
	}

	c.Messages = append(c.Messages, message)

	if len(c.Messages) > s.MaxMessageCount {
		c.Messages = c.Messages[len(c.Messages)-s.MaxMessageCount:]
	}

	return nil
}

func (s *StateData) MessageRemove(guildID string, message *discordgo.Message) error {
	if s == nil {
		return ErrNilState
	}

	return s.messageRemoveByID(guildID, message.ChannelID, message.ID)
}

func (s *StateData) messageRemoveByID(guildID string, channelID string, messageID string) error {
	c, err := s.Channel(guildID, channelID)
	if err != nil {
		return err
	}

	s.Lock()
	defer s.Unlock()

	for i, m := range c.Messages {
		if m.ID == messageID {
			c.Messages = append(c.Messages[:i], c.Messages[i+1:]...)

			return nil
		}
	}

	return ErrStateNotFound
}

func (s *StateData) voiceStateUpdate(update *discordgo.VoiceStateUpdate) error {
	guild, err := s.Guild(update.GuildID)
	if err != nil {
		return err
	}

	s.Lock()
	defer s.Unlock()

	// Handle Leaving Channel
	if update.ChannelID == "" {
		for i, state := range guild.VoiceStates {
			if state.UserID == update.UserID {
				guild.VoiceStates = append(guild.VoiceStates[:i], guild.VoiceStates[i+1:]...)
				return nil
			}
		}
	} else {
		for i, state := range guild.VoiceStates {
			if state.UserID == update.UserID {
				guild.VoiceStates[i] = update.VoiceState
				return nil
			}
		}

		guild.VoiceStates = append(guild.VoiceStates, update.VoiceState)
	}

	return nil
}

func (s *StateData) VoiceState(guildID, userID string) (*discordgo.VoiceState, error) {
	if s == nil {
		return nil, ErrNilState
	}

	guild, err := s.Guild(guildID)
	if err != nil {
		return nil, err
	}

	for _, state := range guild.VoiceStates {
		if state.UserID == userID {
			return state, nil
		}
	}

	return nil, ErrStateNotFound
}

func (s *StateData) Message(guildID string, channelID string, messageID string) (*discordgo.Message, error) {
	if s == nil {
		return nil, ErrNilState
	}

	c, err := s.Channel(guildID, channelID)
	if err != nil {
		return nil, err
	}

	s.RLock()
	defer s.RUnlock()

	for _, m := range c.Messages {
		if m.ID == messageID {
			return m, nil
		}
	}

	return nil, ErrStateNotFound
}

func (s *State) CreateNewShard(shardid int) (err error) {
	if s == nil {
		return ErrNilState
	}

	s.Lock()
	defer s.Unlock()

	s.Shards[shardid] = &StateData{
		MaxMessageCount: 	50,
		guildMap:           make(map[string]*discordgo.Guild),
		memberMap:          make(map[string]map[string]*discordgo.Member),
	}

	return nil
}
