package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	fmt.Println("📋 Invitation Link Generator for DiscMuteBot")
	fmt.Println("============================================")
	fmt.Println()
	fmt.Println("To generate an invitation link, we need your Discord application ID.")
	fmt.Println("You can find this ID at: https://discord.com/developers/applications")
	fmt.Println("1. Select your application")
	fmt.Println("2. The application ID appears at the top of the page")
	fmt.Println()

	// Request application ID
	fmt.Print("Please enter your application ID: ")
	reader := bufio.NewReader(os.Stdin)
	clientID, _ := reader.ReadString('\n')
	clientID = strings.TrimSpace(clientID)

	if clientID == "" {
		fmt.Println("❌ You haven't entered a valid ID. Exiting...")
		return
	}

	// Required permissions:
	// - Administrator (8): Includes all needed permissions
	//requiredPermissions := "8"

	// Generate invitation links (2 different versions)
	simpleLink := fmt.Sprintf("https://discord.com/api/oauth2/authorize?client_id=%s&permissions=8&scope=bot", clientID)
	completeLink := fmt.Sprintf("https://discord.com/api/oauth2/authorize?client_id=%s&permissions=8&scope=bot%%20applications.commands", clientID)

	fmt.Println()
	fmt.Println("✅ Generated invitation links:")
	fmt.Println()
	fmt.Println("Option 1 (recommended):")
	fmt.Println(simpleLink)
	fmt.Println()
	fmt.Println("Option 2 (with application commands):")
	fmt.Println(completeLink)
	fmt.Println()
	fmt.Println("If the first link doesn't work, try the second one.")
	fmt.Println()
	fmt.Println("IMPORTANT: Make sure you have enabled \"Privileged Gateway Intents\" in the Discord Developer portal:")
	fmt.Println("1. Go to https://discord.com/developers/applications")
	fmt.Println("2. Select your application")
	fmt.Println("3. In the \"Bot\" section, enable:")
	fmt.Println("   - SERVER MEMBERS INTENT")
	fmt.Println("   - MESSAGE CONTENT INTENT")
}
