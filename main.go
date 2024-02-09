package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"runtime"
	"strconv"
	"regexp"
)

// Version of the program
const Version = "v1.0.7"

func main() {
	fmt.Println("Program Version:", Version)
// ValidationRules holds validation rules for environment variables
	var ValidationRules = map[string]func(string) bool{
		"Numeric": func(val string) bool {
			// Numeric: Allows only positive numeric values (e.g., "123", "456")
			num, err := strconv.Atoi(val)
			return err == nil && num >= 0
		},
		"Floating": func(val string) bool {
			// Floating: Allows only positive floating-point values (e.g., "3.14", "0.005")
			_, err := strconv.ParseFloat(val, 64)
			if err != nil {
				return false
			}
			// Check if the value contains a decimal point
			decimalPointIndex := strings.Index(val, ".")
			if decimalPointIndex == -1 {
				return false // No decimal point found
			}
			// Check if there are digits after the decimal point
			return decimalPointIndex < len(val)-1
		},
		"TrueFalse": func(val string) bool {
			// TrueFalse: Allows values "True" or "False"
			return val == "True" || val == "False"
		},
		"String": func(val string) bool {
			// String: Allows string values with spaces (e.g., "Hello World", "This is a string")
			return true // No validation needed for string with spaces
		},
		"AlphaDash": func(val string) bool {
			// AlphaDash: Allows only alphanumeric characters and dashes (e.g., "abc123", "test-123")
			return regexp.MustCompile(`^[a-zA-Z0-9_-]+$`).MatchString(val)
		},
		// Add more validation rules as needed
	}


	// Read environment variables
	envVars := map[string]string{
		"Difficulty":                          os.Getenv("DIFFICULTY"),
		"DayTimeSpeedRate":                    os.Getenv("DAY_TIME_SPEED_RATE"),
		"NightTimeSpeedRate":                  os.Getenv("NIGHT_TIME_SPEED_RATE"),
		"ExpRate":                             os.Getenv("EXP_RATE"),
		"PalCaptureRate":                      os.Getenv("PAL_CAPTURE_RATE"),
		"PalSpawnNumRate":                     os.Getenv("PAL_SPAWN_NUM_RATE"),
		"PalDamageRateAttack":                 os.Getenv("PAL_DAMAGE_RATE_ATTACK"),
		"PalDamageRateDefense":                os.Getenv("PAL_DAMAGE_RATE_DEFENSE"),
		"PlayerDamageRateAttack":              os.Getenv("PLAYER_DAMAGE_RATE_ATTACK"),
		"PlayerDamageRateDefense":             os.Getenv("PLAYER_DAMAGE_RATE_DEFENSE"),
		"PlayerStomachDecreaceRate":           os.Getenv("PLAYER_STOMACH_DECREACE_RATE"),
		"PlayerStaminaDecreaceRate":           os.Getenv("PLAYER_STAMINA_DECREACE_RATE"),
		"PlayerAutoHPRegeneRate":              os.Getenv("PLAYER_AUTO_HP_REGENE_RATE"),
		"PlayerAutoHpRegeneRateInSleep":       os.Getenv("PLAYER_AUTO_HP_REGENE_RATE_IN_SLEEP"),
		"PalStomachDecreaceRate":              os.Getenv("PAL_STOMACH_DECREACE_RATE"),
		"PalStaminaDecreaceRate":              os.Getenv("PAL_STAMINA_DECREACE_RATE"),
		"PalAutoHPRegeneRate":                 os.Getenv("PAL_AUTO_HP_REGENE_RATE"),
		"PalAutoHpRegeneRateInSleep":          os.Getenv("PAL_AUTO_HP_REGENE_RATE_IN_SLEEP"),
		"BuildObjectDamageRate":               os.Getenv("BUILD_OBJECT_DAMAGE_RATE"),
		"BuildObjectDeteriorationDamageRate":  os.Getenv("BUILD_OBJECT_DETERIORATION_DAMAGE_RATE"),
		"CollectionDropRate":                  os.Getenv("COLLECTION_DROP_RATE"),
		"CollectionObjectHpRate":              os.Getenv("COLLECTION_OBJECT_HP_RATE"),
		"CollectionObjectRespawnSpeedRate":    os.Getenv("COLLECTION_OBJECT_RESPAWN_SPEED_RATE"),
		"EnemyDropItemRate":                   os.Getenv("ENEMY_DROP_ITEM_RATE"),
		"DeathPenalty":                        os.Getenv("DEATH_PENALTY"),
		"bEnablePlayerToPlayerDamage":         os.Getenv("ENABLE_PLAYER_TO_PLAYER_DAMAGE"),
		"bEnableFriendlyFire":                 os.Getenv("ENABLE_FRIENDLY_FIRE"),
		"bEnableInvaderEnemy":                 os.Getenv("ENABLE_ENEMY"),
		"bActiveUNKO":                         os.Getenv("ACTIVE_UNKO"),
		"bEnableAimAssistPad":                 os.Getenv("ENABLE_AIM_ASSIST_PAD"),
		"bEnableAimAssistKeyboard":            os.Getenv("ENABLE_AIM_ASSIST_KEYBOARD"),
		"DropItemMaxNum":                      os.Getenv("DROP_ITEM_MAX_NUM"),
		"DropItemMaxNum_UNKO":                 os.Getenv("DROP_ITEM_MAX_NUM_UNKO"),
		"BaseCampMaxNum":                      os.Getenv("BASE_CAMP_MAX_NUM"),
		"BaseCampWorkerMaxNum":                os.Getenv("BASE_CAMP_WORKER_MAX_NUM"),
		"DropItemAliveMaxHours":               os.Getenv("DROP_ITEM_ALIVE_MAX_HOURS"),
		"bAutoResetGuildNoOnlinePlayers":      os.Getenv("AUTO_RESET_GUILD_NO_ONLINE_PLAYERS"),
		"AutoResetGuildTimeNoOnlinePlayers":   os.Getenv("AUTO_RESET_GUILD_TIME_NO_ONLINE_PLAYERS"),
		"GuildPlayerMaxNum":                   os.Getenv("GUILD_PLAYER_MAX_NUM"),
		"PalEggDefaultHatchingTime":           os.Getenv("PAL_EGG_DEFAULT_HATCHING_TIME"),
		"WorkSpeedRate":                       os.Getenv("WORK_SPEED_RATE"),
		"bIsMultiplay":                        os.Getenv("IS_MULTIPLAY"),
		"bIsPvP":                              os.Getenv("IS_PVP"),
		"bCanPickupOtherGuildDeathPenaltyDrop":os.Getenv("CAN_PICKUP_OTHER_GUILD_DEATH_PENALTY_DROP"),
		"bEnableNonLoginPenalty":              os.Getenv("ENABLE_NON_LOGIN_PENALTY"),
		"bEnableFastTravel":                   os.Getenv("ENABLE_FAST_TRAVEL"),
		"bIsStartLocationSelectByMap":         os.Getenv("IS_START_LOCATION_SELECT_BY_MAP"),
		"bExistPlayerAfterLogout":             os.Getenv("EXIST_PLAYER_AFTER_LOGOUT"),
		"bEnableDefenseOtherGuildPlayer":      os.Getenv("ENABLE_DEFENSE_OTHER_GUILD_PLAYER"),
		"CoopPlayerMaxNum":                    os.Getenv("COOP_PLAYER_MAX_NUM"),
		"ServerPlayerMaxNum":                  os.Getenv("MAX_PLAYERS"),
		"ServerName":                          os.Getenv("SERVER_NAME"),
		"ServerDescription":                   os.Getenv("SERVER_DESCRIPTION"),
		"ServerPassword":                      os.Getenv("SERVER_PASSWORD"),
		"AdminPassword":                       os.Getenv("ADMIN_PASSWORD"),
		"PublicIP":                            os.Getenv("PUBLIC_IP"),
		"PublicPort":                          os.Getenv("SERVER_PORT"),
		"RCONPort":                            os.Getenv("RCON_PORT"),
		"RCONEnabled":                         os.Getenv("RCON_ENABLE"),
		"bUseAuth":                            os.Getenv("USE_AUTH"),
		"BanListURL":                          os.Getenv("BAN_LIST_URL"),
		"Region":			       os.Getenv("SERVER_REGION"),
		// Add other environment variables and corresponding INI keys here
	}

	// Specify validation rules for each key
	envVarsValidationRules := map[string]string{
		"Difficulty":                     "String", //Difficulty=None,
		"DayTimeSpeedRate":               "Floating", //DayTimeSpeedRate=1.000000,
		"NightTimeSpeedRate":             "Floating", //NightTimeSpeedRate=1.000000,
		"ExpRate":                        "Floating",//ExpRate=1.000000,
		"PalCaptureRate":                 "Floating", //PalCaptureRate=1.000000,
		"PalSpawnNumRate":                "Floating",//PalSpawnNumRate=1.000000,
		"PalDamageRateAttack":            "Floating",//PalDamageRateAttack=1.000000,
		"PalDamageRateDefense":           "Floating",//PalDamageRateDefense=1.000000,
		"PlayerDamageRateAttack":         "Floating",//PlayerDamageRateAttack=1.000000,
		"PlayerDamageRateDefense":        "Floating",//PlayerDamageRateDefense=1.000000,
		"PlayerStomachDecreaceRate":      "Floating",//PlayerStomachDecreaceRate=1.000000,
		"PlayerStaminaDecreaceRate":      "Floating",//PlayerStaminaDecreaceRate=1.000000,
		"PlayerAutoHPRegeneRate":          "Floating",//PlayerAutoHPRegeneRate=1.000000,
		"PlayerAutoHpRegeneRateInSleep":   "Floating",//PlayerAutoHpRegeneRateInSleep=1.000000,
		"PalStaminaDecreaceRate":         "Floating",//PalStaminaDecreaceRate=1.000000,
		"PalStomachDecreaceRate":         "Floating",//PalStomachDecreaceRate=1.000000,
		"PalAutoHPRegeneRate":             "Floating",//PalAutoHPRegeneRate=1.000000,
		"PalAutoHpRegeneRateInSleep":      "Floating",//PalAutoHpRegeneRateInSleep=1.000000,
		"BuildObjectDamageRate":          "Floating",//BuildObjectDamageRate=1.000000,
		"BuildObjectDeteriorationDamageRate": "Floating",//BuildObjectDeteriorationDamageRate=1.000000,
		"CollectionDropRate":             "Floating",//CollectionDropRate=1.000000,
		"CollectionObjectHPRate":         "Floating",//CollectionObjectHpRate=1.000000,
		"CollectionObjectRespawnSpeedRate": "Floating",//CollectionObjectRespawnSpeedRate=1.000000,
		"EnemyDropItemRate":              "Floating",//EnemyDropItemRate=1.000000,
		"DeathPenalty":                   "String",//DeathPenalty=All,
		"bEnablePlayerToPlayerDamage":     "TrueFalse",//bEnablePlayerToPlayerDamage=False,
		"bEnableFriendlyFire":             "TrueFalse",//bEnableFriendlyFire=False,
		"bEnableInvaderEnemy":             "TrueFalse",//bEnableInvaderEnemy=True,
		"bActiveUNKO":                     "TrueFalse",//bActiveUNKO=False,
		"bEnableAimAssistPad":             "TrueFalse",//bEnableAimAssistPad=True,
		"bEnableAimAssistKeyboard":        "TrueFalse",//bEnableAimAssistKeyboard=False,
		"DropItemMaxNum":                 "Numeric",//DropItemMaxNum=3000,
		"DropItemMaxNum_UNKO":            "Numeric",//DropItemMaxNum_UNKO=100,
		"BaseCampMaxNum":                 "Numeric",//BaseCampMaxNum=128,
		"BaseCampWorkerMaxNum":           "Numeric",//BaseCampWorkerMaxNum=15,
		"DropItemAliveMaxHours":          "Floating",//DropItemAliveMaxHours=1.000000,
		"AutoResetGuildTimeNoOnlinePlayers":  "Floating",//AutoResetGuildTimeNoOnlinePlayers=72.000000,
		"bAutoResetGuildNoOnlinePlayers": "TrueFalse",//bAutoResetGuildNoOnlinePlayers=False,
		"GuildPlayerMaxNum":              "Numeric",//GuildPlayerMaxNum=20,
		"PalEggDefaultHatchingTime":      "Floating",//PalEggDefaultHatchingTime=72.000000,
		"WorkSpeedRate":                  "Floating",//WorkSpeedRate=1.000000,
		"bIsMultiplay":                    "TrueFalse",//bIsMultiplay=False,
		"bIsPvP":                          "TrueFalse",//bIsPvP=False,
		"bCanPickupOtherGuildDeathPenaltyDrop": "TrueFalse",//bCanPickupOtherGuildDeathPenaltyDrop=False,
		"bEnableNonLoginPenalty":          "TrueFalse",//bEnableNonLoginPenalty=True,
		"bEnableFastTravel":               "TrueFalse",//bEnableFastTravel=True,
		"bIsStartLocationSelectByMap":     "TrueFalse",//bIsStartLocationSelectByMap=True,
		"bExistPlayerAfterLogout":         "TrueFalse",//bExistPlayerAfterLogout=False,
		"bEnableDefenseOtherGuildPlayer":  "TrueFalse",//bEnableDefenseOtherGuildPlayer=False,
		"CoopPlayerMaxNum":               "Numeric",//CoopPlayerMaxNum=4,
		"ServerPlayerMaxNum":             "Numeric",//ServerPlayerMaxNum=32,
		"ServerName":                     "String",//ServerName="Default Palworld Server",
		"ServerDescription":              "String",//ServerDescription="",
		"ServerPassword":                 "AlphaDash",//ServerPassword="",
		"AdminPassword":                  "AlphaDash",//AdminPassword="",
		"PublicIP":                       "String",//PublicIP="",
		"PublicPort":                     "Numeric",//PublicPort=8211,
		"RCONPort":                       "Numeric",//RCONPort=25575,
		"RCONEnabled":                    "TrueFalse",//RCONEnabled=False,
		"bUseAuth":                        "TrueFalse",//bUseAuth=True,
		"BanListURL":                     "String",//BanListURL="https://api.palworldgame.com/api/banlist.txt"
		"Region":		          "String",//Region="",
		// Add other keys as needed
	}

	// Specify keys for which quotes should be added
	envVarsQuotes := map[string]bool{
		"ServerName":     true,
		"ServerPassword": true,
		"AdminPassword":  true,
		"ServerDescription": true,
		"BanListURL": true,
		"PublicIP": true,
		// Add other keys as needed
	}
	// Determine the operating system
	var osFolder string
	switch runtime.GOOS {
	case "windows":
		osFolder = "WindowsServer"
	case "linux":
		osFolder = "LinuxServer"
	default:
		fmt.Println("Unsupported operating system")
		return
	}

	// Get the absolute path to the INI file
	iniFilePath, err := filepath.Abs(fmt.Sprintf("Pal/Saved/Config/%s/PalWorldSettings.ini", osFolder))
	if err != nil {
		fmt.Printf("Error getting absolute path: %v\n", err)
		return
	}

	// Check if PalWorldSettings.ini exists
	fileInfo, err := os.Stat(iniFilePath)
	if os.IsNotExist(err) {
		// PalWorldSettings.ini does not exist
		// Check if DefaultPalWorldSettings.ini exists in the current directory
		defaultIniPath := "DefaultPalWorldSettings.ini"
		if _, err := os.Stat(defaultIniPath); !os.IsNotExist(err) {
			// DefaultPalWorldSettings.ini exists, so move it to the desired location
			newIniPath := fmt.Sprintf("Pal/Saved/Config/%s/PalWorldSettings.ini", osFolder)
			err := copyFile(defaultIniPath, iniFilePath)
			if err != nil {
				fmt.Printf("Error copying file: %v\n", err)
				return
			}
			fmt.Println("DefaultPalWorldSettings.ini copied to:", newIniPath)
		} else {
			fmt.Println("PalWorldSettings.ini not found and DefaultPalWorldSettings.ini does not exist in the current directory.")
			return // No need to continue if PalWorldSettings.ini doesn't exist and DefaultPalWorldSettings.ini isn't found
		}
	} else if fileInfo.Size() == 0 {
		// PalWorldSettings.ini exists but is empty
		// Copy the default INI file
		defaultIniPath := "DefaultPalWorldSettings.ini"
		newIniPath := fmt.Sprintf("Pal/Saved/Config/%s/PalWorldSettings.ini", osFolder)
		err := copyFile(defaultIniPath, iniFilePath)
		if err != nil {
			fmt.Printf("Error copying file: %v\n", err)
			return
		}
		fmt.Println("DefaultPalWorldSettings.ini copied to:", newIniPath)
	} else {
		fmt.Println("PalWorldSettings.ini found at:", iniFilePath)
	}

	// Read the contents of the original INI file
	iniContent, err := ioutil.ReadFile(iniFilePath)
	if err != nil {
		fmt.Printf("Error reading INI file: %v\n", err)
		return
	}

	// Update values based on environment variables
	for key, value := range envVars {
		// Check if the environment variable exists
		if value == "" && os.Getenv(key) == ""{
			//fmt.Printf("Skipping key: %s because environment variable doesn't exist\n", key)
			continue
		}

		// Skip validation and updating if value is empty but environment variable exists
		if value == "nil" || value == "null" {
			fmt.Printf("Cleared key %s due to %s value.\n", key, value)
			// Set key to empty value in the INI file
			setINIValue(&iniContent, key, "", envVarsQuotes[key])
			continue
		}

		if value != "" {
			// Check if there's a validation rule for the key
			if ruleName, ok := envVarsValidationRules[key]; ok {
				// Check if there's a validation function for the rule name
				if rule, ok := ValidationRules[ruleName]; ok {
					// Validate the value based on the rule
					if !rule(value) {
						fmt.Printf("Validation failed for key: %s, value: %s\n", key, value)
						continue
					}
				} else {
					fmt.Printf("No validation rule found for key: %s\n", key)
				}
			} else {
				fmt.Printf("No validation rule specified for key: %s\n", key)
			}
		}

		// Update the value in the INI file
		fmt.Printf("Updating key: %s with value: %s\n", key, value)
		setINIValue(&iniContent, key, value, envVarsQuotes[key])
	}

	// Write the updated contents back to the INI file
	err = ioutil.WriteFile(iniFilePath, iniContent, 0644)
	if err != nil {
		fmt.Printf("Error writing updated INI file: %v\n", err)
		return
	}

	fmt.Println("INI file updated successfully.")
}

// setINIValue updates the value for the specified key in the INI content.
func setINIValue(content *[]byte, key, value string, addQuotes bool) {
	// Convert content to string for easy manipulation
	contentStr := string(*content)

	// Create the search string for the key
	searchStr := fmt.Sprintf("%s=", key)

	// Find the position of the key in the content
	pos := strings.Index(contentStr, searchStr)
	if pos == -1 {
		// Key not found
		fmt.Printf("Key not found: %s\n", key)
		return
	}

	// Find the end position of the value (comma or end of line)
	endPos := strings.Index(contentStr[pos:], ",")
	if endPos == -1 {
		// If there is no comma, check if it's at the end of the content
		endPos = len(contentStr) - pos
	}

	// If addQuotes is true and the key requires quotes, add quotes around the value
	if addQuotes {
		value = fmt.Sprintf(`"%s"`, value)
	}

	// Calculate the positions in the byte slice
	start := pos + len(searchStr)
	end := pos + endPos

	// Ensure the end position is within the slice bounds
	if end > len(*content) {
		end = len(*content)
	}

	// Update the content slice in place
	*content = append((*content)[:start], append([]byte(value), (*content)[end:]...)...)
}


// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	data, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(dst, data, 0644)
}
