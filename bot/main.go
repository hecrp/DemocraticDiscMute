package main

import (
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"
)

const (
	VOTE_DURATION = 10 * time.Minute
	MUTE_DURATION = 5 * time.Minute
	VOTES_NEEDED = 5
)

type MuteInfo struct {
	MutedBy    map[string]time.Time `json:"muted_by"`
	MuteExpiry time.Time           `json:"mute_expiry"`
	IsGloballyMuted bool           `json:"is_globally_muted"`
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
	// Inicializar el mapa
	muteData.MutedUsers = make(map[string]MuteInfo)
	
	// Cargar configuración
	configFile, err := os.ReadFile("config.json")
	if err != nil {
		log.Fatalf("Error al leer archivo de configuración: %v", err)
	}
	
	err = json.Unmarshal(configFile, &config)
	if err != nil {
		log.Fatalf("Error al parsear archivo de configuración: %v", err)
	}
	
	// Cargar datos existentes si existen
	loadMuteData()
}

func main() {
	if config.Token == "" || config.Token == "TU_TOKEN_AQUI" {
		log.Fatalf("Token no configurado. Edita el archivo config.json.")
	}

	// Configurar logger
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Iniciando bot...")

	dg, err := discordgo.New("Bot " + config.Token)
	if err != nil {
		log.Fatalf("Error al iniciar sesión en Discord: %v", err)
	}

	// Actualizar los intents para incluir mensajes y otros permisos necesarios
	dg.Identify.Intents = discordgo.IntentsGuilds | 
		discordgo.IntentsGuildVoiceStates | 
		discordgo.IntentsGuildMembers |
		discordgo.IntentsGuildMessages | 
		discordgo.IntentsMessageContent

	// Registrar handlers con más logs
	dg.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Bot está listo! Conectado como: %v#%v", s.State.User.Username, s.State.User.Discriminator)
		log.Printf("ID del Bot: %v", s.State.User.ID)
		log.Printf("Servidores a los que pertenece el bot: %d", len(s.State.Guilds))
		for _, g := range s.State.Guilds {
			log.Printf("- Servidor: %s (ID: %s)", g.Name, g.ID)
		}
	})
	
	dg.AddHandler(voiceStateUpdate)
	dg.AddHandler(messageCreate)
	
	// Agregar manejador de errores
	dg.AddHandler(func(s *discordgo.Session, e *discordgo.Connect) {
		log.Println("Conectado a Discord")
	})
	
	dg.AddHandler(func(s *discordgo.Session, e *discordgo.Disconnect) {
		log.Println("Desconectado de Discord")
	})

	err = dg.Open()
	if err != nil {
		log.Fatalf("Error al iniciar sesión en Discord: %v", err)
	}

	fmt.Println("✅ La Ninicracia está en funcionamiento. Presiona Ctrl+C para salir")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop

	err = dg.Close()
	if err != nil {
		return
	}
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignoramos mensajes del bot
	if m.Author.Bot {
		return
	}

	// Log de mensaje recibido
	// log.Printf("Mensaje recibido - Autor: %s, Contenido: %s, Canal: %s", m.Author.Username, m.Content, m.ChannelID)

	// COMANDOS:
	switch {
	case m.Content == "!calladitoping":
		s.ChannelMessageSend(m.ChannelID, "Tú a callar! 🏓")
	case m.Content == "!calladitoservidores":
		// Mostrar servidores a los que pertenece el bot
		var servidores strings.Builder
		servidores.WriteString("🤖 **Servidores a los que pertenezco:**\n\n")
		
		if len(s.State.Guilds) == 0 {
			servidores.WriteString("No estoy en ningún servidor. ¡Invítame usando el enlace generado por la utilidad!\n")
		} else {
			for i, g := range s.State.Guilds {
				guild, err := s.Guild(g.ID)
				if err != nil {
					servidores.WriteString(fmt.Sprintf("%d. **%s** (ID: %s) - Error al obtener detalles\n", i+1, g.Name, g.ID))
					continue
				}
				servidores.WriteString(fmt.Sprintf("%d. **%s** (ID: %s)\n", i+1, guild.Name, guild.ID))
				servidores.WriteString(fmt.Sprintf("   - Miembros: %d\n", guild.MemberCount))
				servidores.WriteString(fmt.Sprintf("   - Región: %s\n", guild.Region))
				
				// Intentar obtener el rol del bot
				member, err := s.GuildMember(g.ID, s.State.User.ID)
				if err != nil {
					servidores.WriteString("   - Roles: Error al obtener roles\n")
				} else if len(member.Roles) == 0 {
					servidores.WriteString("   - Roles: No tengo roles asignados ⚠️\n")
				} else {
					servidores.WriteString("   - Roles: ")
					for j, roleID := range member.Roles {
						role, err := s.State.Role(g.ID, roleID)
						if err != nil {
							servidores.WriteString(fmt.Sprintf("Role %s (error), ", roleID))
						} else {
							servidores.WriteString(fmt.Sprintf("%s", role.Name))
							if j < len(member.Roles)-1 {
								servidores.WriteString(", ")
							}
						}
					}
					servidores.WriteString("\n")
				}
			}
		}
		
		servidores.WriteString("\n**Nota:** Si no ves el bot en el servidor, es posible que necesites:**\n")
		servidores.WriteString("1. Asignar un rol al bot para hacerlo visible\n")
		servidores.WriteString("2. Verificar que el bot tenga permisos para ver el canal donde estás escribiendo\n")
		servidores.WriteString("3. Re-invitar al bot usando el enlace generado por `utils/generar_invitacion.go`\n")
		
		s.ChannelMessageSend(m.ChannelID, servidores.String())
	case m.Content == "!calladitodebug":
		debugInfo := fmt.Sprintf("```\nInformación de Depuración:\n"+
			"- Usuario: %s#%s\n"+
			"- ID de Usuario: %s\n"+
			"- Canal: %s\n"+
			"- Servidor: %s\n"+
			"- Intents habilitados: %d\n"+
			"- Usuarios en mute_data: %d\n```",
			m.Author.Username, m.Author.Discriminator,
			m.Author.ID,
			m.ChannelID,
			m.GuildID,
			s.Identify.Intents,
			len(muteData.MutedUsers))
		s.ChannelMessageSend(m.ChannelID, debugInfo)
	case strings.HasPrefix(m.Content, "!calladito "):
		log.Printf("Comando !calladito detectado")
		if len(m.Mentions) == 0 {
			s.ChannelMessageSend(m.ChannelID, "⚠️ Debes mencionar a un usuario con @ para votar para silenciarlo. Ejemplo: `!calladito @pablito`")
			return
		}
		handleMute(s, m, m.Mentions[0])
	case strings.HasPrefix(m.Content, "!calladitoinfo"):
		fmt.Println("✅ comando muteinfo")
		if len(m.Mentions) == 0 {
			// Mostrar todos los usuarios con votos activos
			handleMuteInfoAll(s, m)
		} else {
			// Mostrar info de un usuario específico
			handleMuteInfo(s, m, m.Mentions[0].ID)
		}
	case m.Content == "!calladitostatus":
		handleMuteStatus(s, m)
	case m.Content == "!calladitohelp":
		handleHelp(s, m)
	}
}

func handleMute(s *discordgo.Session, m *discordgo.MessageCreate, target *discordgo.User) {
	// Verificaciones anti-MRPABLO
	// No permitir votar contra uno mismo
	if target.ID == m.Author.ID {
		s.ChannelMessageSend(m.ChannelID, "⚠️ No puedes votar para mutearte a ti mismo.")
		return
	}

	// No permitir votar contra el bot
	if target.Bot {
		s.ChannelMessageSend(m.ChannelID, "⚠️ No puedes votar para mutear a un bot.")
		return
	}

	muteInfo, exists := muteData.MutedUsers[target.ID]
	if !exists {
		muteInfo = MuteInfo{
			MutedBy: make(map[string]time.Time),
		}
	}

	// Si ya está muteado, informar y salir. Que no se pasen de listos...
	if muteInfo.IsGloballyMuted && time.Now().Before(muteInfo.MuteExpiry) {
		timeLeft := time.Until(muteInfo.MuteExpiry).Round(time.Second)
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("🔇 %s ya está silenciado en canales de voz. El mute terminará en: %s", 
			target.Username, timeLeft))
		return
	}

	// Limpiar votos expirados antes de seguir
	cleanExpiredVotes(&muteInfo)

	// Verificar si el usuario ya ha votado y su voto no ha expirado
	if expiry, hasVoted := muteInfo.MutedBy[m.Author.ID]; hasVoted && time.Now().Before(expiry) {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("⚠️ Ya has votado para silenciar a %s. Tu voto expira en: %s", 
			target.Username, time.Until(expiry).Round(time.Minute).String()))
		return
	}

	// Registrar nuevo voto
	muteInfo.MutedBy[m.Author.ID] = time.Now().Add(VOTE_DURATION)

	// Contar votos activos
	activeVotes := len(muteInfo.MutedBy)

	// Verificar si se alcanzó el umbral de votos y el usuario no está muteado globalmente
	if activeVotes >= VOTES_NEEDED && !muteInfo.IsGloballyMuted {
		// Buscar al usuario en todos los canales de voz del servidor
		guild, err := s.State.Guild(m.GuildID)
		if err != nil {
			log.Printf("Error al obtener información del servidor: %v", err)
			s.ChannelMessageSend(m.ChannelID, "❌ Error al obtener información del servidor.")
			return
		}

		// Intentar silenciar al usuario (solo afecta si está en un canal de voz)
		err = s.GuildMemberMute(m.GuildID, target.ID, true)
		if err != nil {
			log.Printf("Error al silenciar a %s: %v", target.Username, err)
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("❌ Error al silenciar a %s. Es posible que no esté en un canal de voz.", target.Username))
			return
		}

		muteInfo.IsGloballyMuted = true
		muteInfo.MuteExpiry = time.Now().Add(MUTE_DURATION)
		
		// Programar el desmuteo automático
		time.AfterFunc(MUTE_DURATION, func() {
			unmuteUser(s, m.GuildID, target.ID)
		})

		// Verificar si el usuario está actualmente en un canal de voz
		isInVoiceChannel := false
		for _, vs := range guild.VoiceStates {
			if vs.UserID == target.ID {
				isInVoiceChannel = true
				break
			}
		}

		if isInVoiceChannel {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("🔇 %s ha sido muteado en canales de voz por %d minutos.", 
				target.Username, MUTE_DURATION.Minutes()))
		} else {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("🔇 %s será muteado cuando se conecte a un canal de voz. El silenciamiento durará %d minutos.", 
				target.Username, MUTE_DURATION.Minutes()))
		}
	} else {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("✅ Voto registrado contra %s. Votos actuales: %d/%d\nTu voto expira en %d minutos.", 
			target.Username, activeVotes, VOTES_NEEDED, VOTE_DURATION.Minutes()))
	}

	muteData.MutedUsers[target.ID] = muteInfo
	saveMuteData()
}

func handleMuteInfo(s *discordgo.Session, m *discordgo.MessageCreate, targetID string) {
	muteInfo, exists := muteData.MutedUsers[targetID]
	if !exists || len(muteInfo.MutedBy) == 0 {
		s.ChannelMessageSend(m.ChannelID, "📊 No hay votos activos para este usuario.")
		return
	}

	// Obtener info del usuario
	user, err := s.User(targetID)
	if err != nil {
		log.Printf("Error al obtener info del usuario %s: %v", targetID, err)
		user = &discordgo.User{Username: "Usuario desconocido"}
	}

	// Limpiar votos expirados antes de mostrar la información
	cleanExpiredVotes(&muteInfo)
	muteData.MutedUsers[targetID] = muteInfo
	saveMuteData()

	if len(muteInfo.MutedBy) == 0 {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("📊 No hay votos activos para mutear a %s.", user.Username))
		return
	}

	// Crear mensaje con la información
	var msg strings.Builder
	msg.WriteString(fmt.Sprintf("📊 Votos activos para mutear a %s (%d/%d):\n```\n", 
		user.Username, len(muteInfo.MutedBy), VOTES_NEEDED))
	
	for voterID, expiry := range muteInfo.MutedBy {
		// Intentar obtener el username del votante
		username := "Usuario " + voterID
		voter, err := s.User(voterID)
		if err == nil {
			username = voter.Username
		}
		
		timeLeft := time.Until(expiry).Round(time.Second)
		msg.WriteString(fmt.Sprintf("%s (expira en: %s)\n", username, timeLeft))
	}
	msg.WriteString("```")

	if muteInfo.IsGloballyMuted {
		timeLeft := time.Until(muteInfo.MuteExpiry).Round(time.Second)
		if timeLeft > 0 {
			msg.WriteString(fmt.Sprintf("\n🔇 %s está muteado globalmente. Tiempo restante: %s", user.Username, timeLeft))
		}
	}

	s.ChannelMessageSend(m.ChannelID, msg.String())
}

func handleMuteInfoAll(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Verificar si hay usuarios con votos
	if len(muteData.MutedUsers) == 0 {
		s.ChannelMessageSend(m.ChannelID, "📊 No hay votos activos para ningún usuario.")
		return
	}

	// Limpiar votos expirados en todos los usuarios
	for userID, muteInfo := range muteData.MutedUsers {
		cleanExpiredVotes(&muteInfo)
		if len(muteInfo.MutedBy) == 0 {
			delete(muteData.MutedUsers, userID)
		} else {
			muteData.MutedUsers[userID] = muteInfo
		}
	}
	saveMuteData()

	// Verificar de nuevo después de limpiar
	if len(muteData.MutedUsers) == 0 {
		s.ChannelMessageSend(m.ChannelID, "📊 No hay votos activos para ningún usuario.")
		return
	}

	// Crear mensaje con todos los usuarios que tienen votos
	var msg strings.Builder
	msg.WriteString("📊 **Usuarios con votos activos:**\n\n")
	
	for userID, muteInfo := range muteData.MutedUsers {
		// Obtener info del usuario
		username := "Usuario " + userID
		user, err := s.User(userID)
		if err == nil {
			username = user.Username
		}
		
		if muteInfo.IsGloballyMuted {
			timeLeft := time.Until(muteInfo.MuteExpiry).Round(time.Second)
			if timeLeft > 0 {
				msg.WriteString(fmt.Sprintf("🔇 **%s**: Silenciado en voz por %s más - Votos: %d/%d\n", 
					username, timeLeft, len(muteInfo.MutedBy), VOTES_NEEDED))
			} else {
				msg.WriteString(fmt.Sprintf("📊 **%s**: Votos: %d/%d\n", 
					username, len(muteInfo.MutedBy), VOTES_NEEDED))
			}
		} else {
			msg.WriteString(fmt.Sprintf("📊 **%s**: Votos: %d/%d\n", 
				username, len(muteInfo.MutedBy), VOTES_NEEDED))
		}
	}
	
	msg.WriteString("\nUsa `!calladitoinfo @usuario` para ver detalles de un usuario específico.")
	s.ChannelMessageSend(m.ChannelID, msg.String())
}

func handleMuteStatus(s *discordgo.Session, m *discordgo.MessageCreate) {
	var msg strings.Builder
	msg.WriteString("📋 **Estado del sistema de mute:**\n")
	msg.WriteString(fmt.Sprintf("- Votos necesarios: **%d**\n", VOTES_NEEDED))
	msg.WriteString(fmt.Sprintf("- Duración de votos: **%d minutos**\n", VOTE_DURATION.Minutes()))
	msg.WriteString(fmt.Sprintf("- Duración de mute: **%d minutos**\n", MUTE_DURATION.Minutes()))
	
	s.ChannelMessageSend(m.ChannelID, msg.String())
}

func handleHelp(s *discordgo.Session, m *discordgo.MessageCreate) {
	help := "📌 **Comandos de Silenciamiento de Voz:**\n\n" +
		   "**!calladito @usuario** - Vota para silenciar al usuario mencionado en canales de voz\n" +
		   "**!calladitoinfo** - Muestra todos los usuarios con votos activos\n" +
		   "**!calladitoinfo @usuario** - Muestra los votos para un usuario específico\n" +
		   "**!calladitostatus** - Muestra la configuración del sistema de silenciamiento\n" +
		   "**!calladitohelp** - Muestra este mensaje de ayuda\n\n" +
		   fmt.Sprintf("Se necesitan **%d votos** para silenciar a un usuario por **%d minutos**. El silenciamiento solo afecta a los canales de voz.", 
			   VOTES_NEEDED, MUTE_DURATION.Minutes())
	
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
		log.Printf("Error al quitar silenciamiento a %s: %v", userID, err)
		return
	}

	// Obtener info del usuario
	username := "Usuario " + userID
	user, err := s.User(userID)
	if err == nil {
		username = user.Username
	}

	// Notificar en un canal de log (no tenemos m.ChannelID aquí, así que es mejor no enviarlo o implementar otra solución)
	log.Printf("🔊 El silenciamiento de voz de %s ha finalizado", username)

	muteInfo.IsGloballyMuted = false
	muteData.MutedUsers[userID] = muteInfo
	saveMuteData()
}

func voiceStateUpdate(s *discordgo.Session, v *discordgo.VoiceStateUpdate) {
	// Verificar si hay usuarios muteados
	muteInfo, exists := muteData.MutedUsers[v.UserID]
	if !exists || !muteInfo.IsGloballyMuted {
		return
	}

	// Si el usuario está muteado y el mute no ha expirado, asegurarse de que siga muteado cuando se une a un canal de voz
	if time.Now().Before(muteInfo.MuteExpiry) {
		// Solo aplicar mute si el usuario se ha unido a un canal de voz (v.ChannelID no está vacío)
		if v.ChannelID != "" {
			err := s.GuildMemberMute(v.GuildID, v.UserID, true)
			if err != nil {
				log.Printf("Error al mantener silenciamiento de %s: %v", v.UserID, err)
			} else {
				log.Printf("Usuario silenciado %s se unió a un canal de voz, manteniendo silenciamiento", v.UserID)
			}
		}
	} else {
		// Si el mute ha expirado, desmutear al usuario
		unmuteUser(s, v.GuildID, v.UserID)
	}
}

func loadMuteData() {
	data, err := os.ReadFile(muteFile)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("Error al leer archivo de mutes: %v", err)
		}
		return
	}

	err = json.Unmarshal(data, &muteData)
	if err != nil {
		log.Printf("Error al deserializar datos de mute: %v", err)
	}
}

func saveMuteData() {
	data, err := json.MarshalIndent(muteData, "", "    ")
	if err != nil {
		log.Printf("Error al serializar datos de mute: %v", err)
		return
	}

	err = os.WriteFile(muteFile, data, 0644)
	if err != nil {
		log.Printf("Error al guardar archivo de mutes: %v", err)
	}
}
