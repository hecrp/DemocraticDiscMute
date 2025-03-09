package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

const (
	VOTE_DURATION = 10 * time.Minute
	MUTE_DURATION = 5 * time.Minute
	VOTES_NEEDED  = 5
)

type MuteInfo struct {
	MutedBy         map[string]time.Time `json:"muted_by"`
	MuteExpiry      time.Time            `json:"mute_expiry"`
	IsGloballyMuted bool                 `json:"is_globally_muted"`
}

type MuteData struct {
	MutedUsers map[string]MuteInfo `json:"muted_users"`
}

var (
	muteData MuteData
	muteFile = "mute_data.json"
	config   struct {
		Token string `json:"token"`
	}
)

func init() {
	// Initialize the map
	muteData.MutedUsers = make(map[string]MuteInfo)

	// Load configuration
	configFile, err := os.ReadFile("config.json")
	if err != nil {
		log.Fatalf("Error reading configuration file: %v", err)
	}

	err = json.Unmarshal(configFile, &config)
	if err != nil {
		log.Fatalf("Error parsing configuration file: %v", err)
	}

	// Load existing data if it exists
	loadMuteData()
}

func main() {
	if config.Token == "" || config.Token == "TU_TOKEN_AQUI" {
		log.Fatalf("Token not configured. Edit the config.json file.")
	}

	// Configure logger
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting bot...")

	dg, err := discordgo.New("Bot " + config.Token)
	if err != nil {
		log.Fatalf("Error logging in to Discord: %v", err)
	}

	// Update intents to include necessary permissions
	dg.Identify.Intents = discordgo.IntentsGuilds |
		discordgo.IntentsGuildVoiceStates |
		discordgo.IntentsGuildMembers |
		discordgo.IntentsGuildMessages |
		discordgo.IntentsMessageContent

	// Register handlers with more logs
	dg.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Bot is ready! Connected as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
		log.Printf("Bot ID: %v", s.State.User.ID)
		log.Printf("Servers the bot belongs to: %d", len(s.State.Guilds))
		for _, g := range s.State.Guilds {
			log.Printf("- Server: %s (ID: %s)", g.Name, g.ID)
		}
	})

	dg.AddHandler(voiceStateUpdate)
	dg.AddHandler(messageCreate)

	// Add error handler
	dg.AddHandler(func(s *discordgo.Session, e *discordgo.Connect) {
		log.Println("Connected to Discord")
	})

	dg.AddHandler(func(s *discordgo.Session, e *discordgo.Disconnect) {
		log.Println("Disconnected from Discord")
	})

	err = dg.Open()
	if err != nil {
		log.Fatalf("Error logging in to Discord: %v", err)
	}

	fmt.Println("✅ The Ninicracia is in operation. Press Ctrl+C to exit")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop

	err = dg.Close()
	if err != nil {
		return
	}
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore messages from the bot
	if m.Author.Bot {
		return
	}

	// Log received message
	// log.Printf("Message received - Author: %s, Content: %s, Channel: %s", m.Author.Username, m.Content, m.ChannelID)

	// COMMANDS:
	switch {
	case m.Content == "!ping":
		s.ChannelMessageSend(m.ChannelID, "Pong! 🏓")
	case m.Content == "!servers":
		// Show servers to which the bot belongs
		var servidores strings.Builder
		servidores.WriteString("🤖 **Servers to which I belong:**\n\n")

		if len(s.State.Guilds) == 0 {
			servidores.WriteString("I'm not in any server. Invite me using the generated link!\n")
		} else {
			for i, g := range s.State.Guilds {
				guild, err := s.Guild(g.ID)
				if err != nil {
					servidores.WriteString(fmt.Sprintf("%d. **%s** (ID: %s) - Error getting details\n", i+1, g.Name, g.ID))
					continue
				}
				servidores.WriteString(fmt.Sprintf("%d. **%s** (ID: %s)\n", i+1, guild.Name, guild.ID))
				servidores.WriteString(fmt.Sprintf("   - Members: %d\n", guild.MemberCount))
				servidores.WriteString(fmt.Sprintf("   - Region: %s\n", guild.Region))

				// Try to get the bot's role
				member, err := s.GuildMember(g.ID, s.State.User.ID)
				if err != nil {
					servidores.WriteString("   - Roles: Error getting roles\n")
				} else if len(member.Roles) == 0 {
					servidores.WriteString("   - Roles: I don't have assigned roles ⚠️\n")
				} else {
					servidores.WriteString("   - Roles: ")
					for j, roleID := range member.Roles {
						role, err := s.State.Role(g.ID, roleID)
						if err != nil {
							servidores.WriteString(fmt.Sprintf("Role %s (error), ", roleID))
						} else {
							servidores.WriteString(role.Name)
							if j < len(member.Roles)-1 {
								servidores.WriteString(", ")
							}
						}
					}
					servidores.WriteString("\n")
				}
				servidores.WriteString("\n")
			}
		}

		servidores.WriteString("\n**Note:** If you don't see the bot in the server, you may need to:**\n")
		servidores.WriteString("1. Assign a role to the bot to make it visible\n")
		servidores.WriteString("2. Verify that the bot has permissions to see the channel where you're writing\n")
		servidores.WriteString("3. Re-invite the bot using the generated link from `utils/generate_invitation.go`\n")

		s.ChannelMessageSend(m.ChannelID, servidores.String())
	case m.Content == "!debug":
		debugInfo := fmt.Sprintf("```\nDebug Information:\n"+
			"- User: %s#%s\n"+
			"- User ID: %s\n"+
			"- Channel: %s\n"+
			"- Server: %s\n"+
			"- Enabled Intents: %d\n"+
			"- Users in mute_data: %d\n```",
			m.Author.Username, m.Author.Discriminator,
			m.Author.ID,
			m.ChannelID,
			m.GuildID,
			s.Identify.Intents,
			len(muteData.MutedUsers))
		s.ChannelMessageSend(m.ChannelID, debugInfo)
	case strings.HasPrefix(m.Content, "!mute "):
		log.Printf("!mute command detected")
		if len(m.Mentions) == 0 {
			s.ChannelMessageSend(m.ChannelID, "⚠️ You must mention a user with @ to vote to mute them. Example: `!mute @pablito`")
			return
		}
		handleMute(s, m, m.Mentions[0])
	case strings.HasPrefix(m.Content, "!muteinfo"):
		fmt.Println("✅ muteinfo command")
		if len(m.Mentions) == 0 {
			// Show all users with active votes
			handleMuteInfoAll(s, m)
		} else {
			// Show info of a specific user
			handleMuteInfo(s, m, m.Mentions[0].ID)
		}
	case m.Content == "!mutestatus":
		handleMuteStatus(s, m)
	case m.Content == "!help":
		handleHelp(s, m)
	case strings.HasPrefix(m.Content, "!clean"):
		if len(m.Mentions) == 0 {
			s.ChannelMessageSend(m.ChannelID, "❌ Please mention a user to clear their votes. Example: `!clean @user`")
			return
		}

		// Verify if the message author is an administrator
		member, err := s.GuildMember(m.GuildID, m.Author.ID)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "❌ Error verifying administrator permissions")
			log.Printf("Error verifying administrator permissions: %v", err)
			return
		}

		// Verify if the user has administrator permissions
		hasAdminPerms := false

		// Get server roles
		guildRoles, err := s.GuildRoles(m.GuildID)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "❌ Error verifying administrator permissions")
			log.Printf("Error getting server roles: %v", err)
			return
		}

		for _, roleID := range member.Roles {
			for _, guildRole := range guildRoles {
				if guildRole.ID == roleID {
					// Verify if the role has administrator permissions
					if guildRole.Permissions&discordgo.PermissionAdministrator == discordgo.PermissionAdministrator {
						hasAdminPerms = true
						break
					}
				}
			}
			if hasAdminPerms {
				break
			}
		}

		if !hasAdminPerms {
			s.ChannelMessageSend(m.ChannelID, "❌ You don't have administrator permissions to use this command")
			return
		}

		// Process the mention
		target := m.Mentions[0]
		handleClean(s, m, target)
	}
}

func handleMute(s *discordgo.Session, m *discordgo.MessageCreate, target *discordgo.User) {
	// Anti-MRPABLO checks
	// Don't allow voting against oneself
	if target.ID == m.Author.ID {
		s.ChannelMessageSend(m.ChannelID, "⚠️ You can't vote to mute yourself.")
		return
	}

	// Don't allow voting against the bot
	if target.Bot {
		s.ChannelMessageSend(m.ChannelID, "⚠️ You can't vote to mute a bot.")
		return
	}

	muteInfo, exists := muteData.MutedUsers[target.ID]
	if !exists {
		muteInfo = MuteInfo{
			MutedBy: make(map[string]time.Time),
		}
	}

	// If already muted, inform and exit. Don't get ahead of yourself...
	if muteInfo.IsGloballyMuted && time.Now().Before(muteInfo.MuteExpiry) {
		timeLeft := time.Until(muteInfo.MuteExpiry).Round(time.Second)
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("🔇 %s is already muted in voice channels. The mute will end in: %s",
			target.Username, timeLeft))
		return
	}

	// Clean expired votes before proceeding
	cleanExpiredVotes(&muteInfo)

	// Verify if the user has already voted and their vote hasn't expired
	if expiry, hasVoted := muteInfo.MutedBy[m.Author.ID]; hasVoted && time.Now().Before(expiry) {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("⚠️ You've already voted to mute %s. Your vote expires in: %s",
			target.Username, time.Until(expiry).Round(time.Minute).String()))
		return
	}

	// Register new vote
	muteInfo.MutedBy[m.Author.ID] = time.Now().Add(VOTE_DURATION)

	// Count active votes
	activeVotes := len(muteInfo.MutedBy)

	// Register vote in log
	logAction("VOTE", m.Author.Username, target.Username, activeVotes, m.GuildID)

	// Verify if the threshold of votes is reached and the user isn't globally muted
	if activeVotes >= VOTES_NEEDED && !muteInfo.IsGloballyMuted {
		// Find the user in all voice channels of the server
		guild, err := s.State.Guild(m.GuildID)
		if err != nil {
			log.Printf("Error getting server information: %v", err)
			s.ChannelMessageSend(m.ChannelID, "❌ Error getting server information.")
			return
		}

		// Try to mute the user (only affects if they're in a voice channel)
		err = s.GuildMemberMute(m.GuildID, target.ID, true)
		if err != nil {
			log.Printf("Error muting %s: %v", target.Username, err)
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("❌ Error muting %s. It's possible they're not in a voice channel.", target.Username))
			return
		}

		muteInfo.IsGloballyMuted = true
		muteInfo.MuteExpiry = time.Now().Add(MUTE_DURATION)

		// Register mute in log
		logAction("MUTE", m.Author.Username, target.Username, activeVotes, m.GuildID)

		// Schedule automatic unmute
		time.AfterFunc(MUTE_DURATION, func() {
			unmuteUser(s, m.GuildID, target.ID)
		})

		// Verify if the user is currently in a voice channel
		isInVoiceChannel := false
		for _, vs := range guild.VoiceStates {
			if vs.UserID == target.ID {
				isInVoiceChannel = true
				break
			}
		}

		if isInVoiceChannel {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("🔇 %s has been muted in voice channels for %d minutes.",
				target.Username, int(MUTE_DURATION.Minutes())))
		} else {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("🔇 %s will be muted when they join a voice channel. The mute will last %d minutes.",
				target.Username, int(MUTE_DURATION.Minutes())))
		}
	} else {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("✅ Vote registered against %s. Current votes: %d/%d\nYour vote expires in %d minutes.",
			target.Username, activeVotes, VOTES_NEEDED, int(VOTE_DURATION.Minutes())))
	}

	muteData.MutedUsers[target.ID] = muteInfo
	saveMuteData()
}

func handleMuteInfo(s *discordgo.Session, m *discordgo.MessageCreate, targetID string) {
	muteInfo, exists := muteData.MutedUsers[targetID]
	if !exists || len(muteInfo.MutedBy) == 0 {
		s.ChannelMessageSend(m.ChannelID, "📊 No active votes for this user.")
		return
	}

	// Get user info
	user, err := s.User(targetID)
	if err != nil {
		log.Printf("Error getting info for user %s: %v", targetID, err)
		user = &discordgo.User{Username: "Unknown user"}
	}

	// Clean expired votes before showing information
	cleanExpiredVotes(&muteInfo)
	muteData.MutedUsers[targetID] = muteInfo
	saveMuteData()

	if len(muteInfo.MutedBy) == 0 {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("📊 No active votes to mute %s.", user.Username))
		return
	}

	// Create message with information
	var msg strings.Builder
	msg.WriteString(fmt.Sprintf("📊 Active votes to mute %s (%d/%d):\n```\n",
		user.Username, len(muteInfo.MutedBy), VOTES_NEEDED))

	for voterID, expiry := range muteInfo.MutedBy {
		// Try to get the username of the voter
		username := "User " + voterID
		voter, err := s.User(voterID)
		if err == nil {
			username = voter.Username
		}

		timeLeft := time.Until(expiry).Round(time.Second)
		msg.WriteString(fmt.Sprintf("%s (expires in: %s)\n", username, timeLeft))
	}
	msg.WriteString("```")

	if muteInfo.IsGloballyMuted {
		timeLeft := time.Until(muteInfo.MuteExpiry).Round(time.Second)
		if timeLeft > 0 {
			msg.WriteString(fmt.Sprintf("\n🔇 %s is globally muted. Time remaining: %s", user.Username, timeLeft))
		}
	}

	s.ChannelMessageSend(m.ChannelID, msg.String())
}

func handleMuteInfoAll(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Verify if there are users with votes
	if len(muteData.MutedUsers) == 0 {
		s.ChannelMessageSend(m.ChannelID, "📊 No active votes for any user.")
		return
	}

	// Clean expired votes in all users
	for userID, muteInfo := range muteData.MutedUsers {
		cleanExpiredVotes(&muteInfo)
		if len(muteInfo.MutedBy) == 0 {
			delete(muteData.MutedUsers, userID)
		} else {
			muteData.MutedUsers[userID] = muteInfo
		}
	}
	saveMuteData()

	// Verify again after cleaning
	if len(muteData.MutedUsers) == 0 {
		s.ChannelMessageSend(m.ChannelID, "📊 No active votes for any user.")
		return
	}

	// Create message with all users who have votes
	var msg strings.Builder
	msg.WriteString("📊 **Users with active votes:**\n\n")

	for userID, muteInfo := range muteData.MutedUsers {
		// Get user info
		username := "User " + userID
		user, err := s.User(userID)
		if err == nil {
			username = user.Username
		}

		if muteInfo.IsGloballyMuted {
			timeLeft := time.Until(muteInfo.MuteExpiry).Round(time.Second)
			if timeLeft > 0 {
				msg.WriteString(fmt.Sprintf("🔇 **%s**: Muted in voice for %s more - Votes: %d/%d\n",
					username, timeLeft, len(muteInfo.MutedBy), VOTES_NEEDED))
			} else {
				msg.WriteString(fmt.Sprintf("📊 **%s**: Votes: %d/%d\n",
					username, len(muteInfo.MutedBy), VOTES_NEEDED))
			}
		} else {
			msg.WriteString(fmt.Sprintf("📊 **%s**: Votes: %d/%d\n",
				username, len(muteInfo.MutedBy), VOTES_NEEDED))
		}
	}

	msg.WriteString("\nUse `!muteinfo @user` to see details of a specific user.")
	s.ChannelMessageSend(m.ChannelID, msg.String())
}

func handleMuteStatus(s *discordgo.Session, m *discordgo.MessageCreate) {
	var msg strings.Builder
	msg.WriteString("📋 **Mute system status:**\n")
	msg.WriteString(fmt.Sprintf("- Votes needed: **%d**\n", VOTES_NEEDED))
	msg.WriteString(fmt.Sprintf("- Vote duration: **%d minutes**\n", int(VOTE_DURATION.Minutes())))
	msg.WriteString(fmt.Sprintf("- Mute duration: **%d minutes**\n", int(MUTE_DURATION.Minutes())))

	s.ChannelMessageSend(m.ChannelID, msg.String())
}

func handleHelp(s *discordgo.Session, m *discordgo.MessageCreate) {
	help := "📌 **Voice Mute Commands:**\n\n" +
		"**!mute @user** - Vote to mute the mentioned user in voice channels\n" +
		"**!muteinfo** - Show all users with active votes\n" +
		"**!muteinfo @user** - Show votes for a specific user\n" +
		"**!mutestatus** - Show mute system configuration\n" +
		"**!clean @user** - (Only administrators) Remove all votes against a user\n" +
		"**!help** - Show this help message\n\n" +
		fmt.Sprintf("**%d votes** are needed to mute a user for **%d minutes**. The mute only affects voice channels.",
			VOTES_NEEDED, int(MUTE_DURATION.Minutes()))

	s.ChannelMessageSend(m.ChannelID, help)
}

func cleanExpiredVotes(muteInfo *MuteInfo) {
	now := time.Now()
	for user, expiry := range muteInfo.MutedBy {
		if now.After(expiry) {
			delete(muteInfo.MutedBy, user)
		}
	}
}

func unmuteUser(s *discordgo.Session, guildID string, userID string) {
	muteInfo, exists := muteData.MutedUsers[userID]
	if !exists || !muteInfo.IsGloballyMuted {
		return
	}

	err := s.GuildMemberMute(guildID, userID, false)
	if err != nil {
		log.Printf("Error unmuting user %s: %v", userID, err)
		return
	}

	// Update user status
	muteInfo.IsGloballyMuted = false
	muteData.MutedUsers[userID] = muteInfo
	saveMuteData()

	// Register action in log
	// Get user name
	user, err := s.User(userID)
	username := userID // Default to ID if we can't get the name
	if err == nil {
		username = user.Username
	}

	logAction("UNMUTE", "System", username, 0, guildID)

	// Log for debug
	log.Printf("User %s unmuted automatically", userID)
}

func voiceStateUpdate(s *discordgo.Session, v *discordgo.VoiceStateUpdate) {
	// Verify if there are muted users
	muteInfo, exists := muteData.MutedUsers[v.UserID]
	if !exists || !muteInfo.IsGloballyMuted {
		return
	}

	// If the user is muted and the mute hasn't expired, ensure they're muted when they join a voice channel
	if time.Now().Before(muteInfo.MuteExpiry) {
		// Only apply mute if the user has joined a voice channel (v.ChannelID isn't empty)
		if v.ChannelID != "" {
			err := s.GuildMemberMute(v.GuildID, v.UserID, true)
			if err != nil {
				log.Printf("Error maintaining mute for %s: %v", v.UserID, err)
			} else {
				log.Printf("User %s muted in voice channel, maintaining mute", v.UserID)
			}
		}
	} else {
		// If mute has expired, unmute the user
		unmuteUser(s, v.GuildID, v.UserID)
	}
}

func loadMuteData() {
	data, err := os.ReadFile(muteFile)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("Error reading mute file: %v", err)
		}
		return
	}

	err = json.Unmarshal(data, &muteData)
	if err != nil {
		log.Printf("Error deserializing mute data: %v", err)
	}
}

func saveMuteData() {
	data, err := json.MarshalIndent(muteData, "", "    ")
	if err != nil {
		log.Printf("Error serializing mute data: %v", err)
		return
	}

	err = os.WriteFile(muteFile, data, 0644)
	if err != nil {
		log.Printf("Error saving mute file: %v", err)
	}
}

func handleClean(s *discordgo.Session, m *discordgo.MessageCreate, target *discordgo.User) {
	// Verify if the user is in the mute list
	_, exists := muteData.MutedUsers[target.ID]
	if !exists {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(" The user %s doesn't have active votes", target.Username))
		return
	}

	// If the user is muted, unmute
	if muteData.MutedUsers[target.ID].IsGloballyMuted {
		err := s.GuildMemberMute(m.GuildID, target.ID, false)
		if err != nil {
			log.Printf("Error unmuting %s: %v", target.Username, err)
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("⚠️ Error unmuting %s", target.Username))
		} else {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("🔊 %s has been unmuted by an administrator", target.Username))
		}
	}

	// Remove user from mute list
	delete(muteData.MutedUsers, target.ID)
	saveMuteData()

	// Register action in log
	logAction("CLEAN", m.Author.Username, target.Username, 0, m.GuildID)

	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("🧹 All votes against %s have been removed", target.Username))
}

// Logging system
func logAction(actionType, initiator, target string, currentVotes int, guildID string) {
	// Create logs directory if it doesn't exist
	err := os.MkdirAll("logs", 0755)
	if err != nil {
		log.Printf("Error creating logs directory: %v", err)
		return
	}

	// Filename based on current date
	currentDate := time.Now().Format("2006-01-02")
	logFile := fmt.Sprintf("logs/%s.csv", currentDate)

	// Verify if the file exists
	fileExists := false
	if _, err := os.Stat(logFile); err == nil {
		fileExists = true
	}

	// Open file in append mode
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Error opening log file: %v", err)
		return
	}
	defer file.Close()

	// Write header if the file is new
	writer := csv.NewWriter(file)
	defer writer.Flush()

	if !fileExists {
		header := []string{"Timestamp", "ActionType", "Initiator", "Target", "CurrentVotes", "GuildID"}
		err = writer.Write(header)
		if err != nil {
			log.Printf("Error writing log header: %v", err)
			return
		}
	}

	// Write record
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	record := []string{timestamp, actionType, initiator, target, fmt.Sprintf("%d", currentVotes), guildID}
	err = writer.Write(record)
	if err != nil {
		log.Printf("Error writing log record: %v", err)
		return
	}
}
