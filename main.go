package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

// Version of the program
const Version = "v1.0.14"

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
		"Difficulty":                           "DIFFICULTY",
		"DayTimeSpeedRate":                     "DAY_TIME_SPEED_RATE",
		"NightTimeSpeedRate":                   "NIGHT_TIME_SPEED_RATE",
		"ExpRate":                              "EXP_RATE",
		"PalCaptureRate":                       "PAL_CAPTURE_RATE",
		"PalSpawnNumRate":                      "PAL_SPAWN_NUM_RATE",
		"PalDamageRateAttack":                  "PAL_DAMAGE_RATE_ATTACK",
		"PalDamageRateDefense":                 "PAL_DAMAGE_RATE_DEFENSE",
		"PlayerDamageRateAttack":               "PLAYER_DAMAGE_RATE_ATTACK",
		"PlayerDamageRateDefense":              "PLAYER_DAMAGE_RATE_DEFENSE",
		"PlayerStomachDecreaceRate":            "PLAYER_STOMACH_DECREACE_RATE",
		"PlayerStaminaDecreaceRate":            "PLAYER_STAMINA_DECREACE_RATE",
		"PlayerAutoHPRegeneRate":               "PLAYER_AUTO_HP_REGENE_RATE",
		"PlayerAutoHpRegeneRateInSleep":        "PLAYER_AUTO_HP_REGENE_RATE_IN_SLEEP",
		"PalStomachDecreaceRate":               "PAL_STOMACH_DECREACE_RATE",
		"PalStaminaDecreaceRate":               "PAL_STAMINA_DECREACE_RATE",
		"PalAutoHPRegeneRate":                  "PAL_AUTO_HP_REGENE_RATE",
		"PalAutoHpRegeneRateInSleep":           "PAL_AUTO_HP_REGENE_RATE_IN_SLEEP",
		"BuildObjectDamageRate":                "BUILD_OBJECT_DAMAGE_RATE",
		"BuildObjectDeteriorationDamageRate":   "BUILD_OBJECT_DETERIORATION_DAMAGE_RATE",
		"CollectionDropRate":                   "COLLECTION_DROP_RATE",
		"CollectionObjectHpRate":               "COLLECTION_OBJECT_HP_RATE",
		"CollectionObjectRespawnSpeedRate":     "COLLECTION_OBJECT_RESPAWN_SPEED_RATE",
		"EnemyDropItemRate":                    "ENEMY_DROP_ITEM_RATE",
		"DeathPenalty":                         "DEATH_PENALTY",
		"bEnablePlayerToPlayerDamage":          "ENABLE_PLAYER_TO_PLAYER_DAMAGE",
		"bEnableFriendlyFire":                  "ENABLE_FRIENDLY_FIRE",
		"bEnableInvaderEnemy":                  "ENABLE_ENEMY",
		"bActiveUNKO":                          "ACTIVE_UNKO",
		"bEnableAimAssistPad":                  "ENABLE_AIM_ASSIST_PAD",
		"bEnableAimAssistKeyboard":             "ENABLE_AIM_ASSIST_KEYBOARD",
		"DropItemMaxNum":                       "DROP_ITEM_MAX_NUM",
		"DropItemMaxNum_UNKO":                  "DROP_ITEM_MAX_NUM_UNKO",
		"BaseCampMaxNum":                       "BASE_CAMP_MAX_NUM",
		"BaseCampWorkerMaxNum":                 "BASE_CAMP_WORKER_MAX_NUM",
		"DropItemAliveMaxHours":                "DROP_ITEM_ALIVE_MAX_HOURS",
		"bAutoResetGuildNoOnlinePlayers":       "AUTO_RESET_GUILD_NO_ONLINE_PLAYERS",
		"AutoResetGuildTimeNoOnlinePlayers":    "AUTO_RESET_GUILD_TIME_NO_ONLINE_PLAYERS",
		"GuildPlayerMaxNum":                    "GUILD_PLAYER_MAX_NUM",
		"PalEggDefaultHatchingTime":            "PAL_EGG_DEFAULT_HATCHING_TIME",
		"WorkSpeedRate":                        "WORK_SPEED_RATE",
		"bIsMultiplay":                         "IS_MULTIPLAY",
		"bIsPvP":                               "IS_PVP",
		"bCanPickupOtherGuildDeathPenaltyDrop": "CAN_PICKUP_OTHER_GUILD_DEATH_PENALTY_DROP",
		"bEnableNonLoginPenalty":               "ENABLE_NON_LOGIN_PENALTY",
		"bEnableFastTravel":                    "ENABLE_FAST_TRAVEL",
		"bIsStartLocationSelectByMap":          "IS_START_LOCATION_SELECT_BY_MAP",
		"bExistPlayerAfterLogout":              "EXIST_PLAYER_AFTER_LOGOUT",
		"bEnableDefenseOtherGuildPlayer":       "ENABLE_DEFENSE_OTHER_GUILD_PLAYER",
		"CoopPlayerMaxNum":                     "COOP_PLAYER_MAX_NUM",
		"ServerPlayerMaxNum":                   "MAX_PLAYERS",
		"ServerName":                           "SERVER_NAME",
		"ServerDescription":                    "SERVER_DESCRIPTION",
		"ServerPassword":                       "SERVER_PASSWORD",
		"AdminPassword":                        "ADMIN_PASSWORD",
		"PublicPort":                           "SERVER_PORT",
		"RCONPort":                             "RCON_PORT",
		"RCONEnabled":                          "RCON_ENABLE",
		"bUseAuth":                             "USE_AUTH",
		"BanListURL":                           "BAN_LIST_URL",
		"Region":                               "SERVER_REGION",
		"bShowPlayerList":                      "SHOW_PLAYER_LIST",
		"RESTAPIEnabled":                       "REST_API_ENABLED",
		"RESTAPIPort":                          "REST_API_PORT",
		"AllowConnectPlatform":                 "ALLOW_CONNECT_PLATFORM",
		"bIsUseBackupSaveData":                 "USE_BACKUP_SAVE_DATA",
		"LogFormatType":                        "LOG_FORMAT_TYPE",
		// Add other environment variables and corresponding INI keys here
	}

	// Assign value to "PublicIP" key using the function getIPAddressKey()
	envVars["PublicIP"] = getIPAddressKey()

	// Specify validation rules for each key
	envVarsValidationRules := map[string]string{
		"Difficulty":                           "String",    //Difficulty=None,
		"DayTimeSpeedRate":                     "Floating",  //DayTimeSpeedRate=1.000000,
		"NightTimeSpeedRate":                   "Floating",  //NightTimeSpeedRate=1.000000,
		"ExpRate":                              "Floating",  //ExpRate=1.000000,
		"PalCaptureRate":                       "Floating",  //PalCaptureRate=1.000000,
		"PalSpawnNumRate":                      "Floating",  //PalSpawnNumRate=1.000000,
		"PalDamageRateAttack":                  "Floating",  //PalDamageRateAttack=1.000000,
		"PalDamageRateDefense":                 "Floating",  //PalDamageRateDefense=1.000000,
		"PlayerDamageRateAttack":               "Floating",  //PlayerDamageRateAttack=1.000000,
		"PlayerDamageRateDefense":              "Floating",  //PlayerDamageRateDefense=1.000000,
		"PlayerStomachDecreaceRate":            "Floating",  //PlayerStomachDecreaceRate=1.000000,
		"PlayerStaminaDecreaceRate":            "Floating",  //PlayerStaminaDecreaceRate=1.000000,
		"PlayerAutoHPRegeneRate":               "Floating",  //PlayerAutoHPRegeneRate=1.000000,
		"PlayerAutoHpRegeneRateInSleep":        "Floating",  //PlayerAutoHpRegeneRateInSleep=1.000000,
		"PalStaminaDecreaceRate":               "Floating",  //PalStaminaDecreaceRate=1.000000,
		"PalStomachDecreaceRate":               "Floating",  //PalStomachDecreaceRate=1.000000,
		"PalAutoHPRegeneRate":                  "Floating",  //PalAutoHPRegeneRate=1.000000,
		"PalAutoHpRegeneRateInSleep":           "Floating",  //PalAutoHpRegeneRateInSleep=1.000000,
		"BuildObjectDamageRate":                "Floating",  //BuildObjectDamageRate=1.000000,
		"BuildObjectDeteriorationDamageRate":   "Floating",  //BuildObjectDeteriorationDamageRate=1.000000,
		"CollectionDropRate":                   "Floating",  //CollectionDropRate=1.000000,
		"CollectionObjectHPRate":               "Floating",  //CollectionObjectHpRate=1.000000,
		"CollectionObjectRespawnSpeedRate":     "Floating",  //CollectionObjectRespawnSpeedRate=1.000000,
		"EnemyDropItemRate":                    "Floating",  //EnemyDropItemRate=1.000000,
		"DeathPenalty":                         "String",    //DeathPenalty=All,
		"bEnablePlayerToPlayerDamage":          "TrueFalse", //bEnablePlayerToPlayerDamage=False,
		"bEnableFriendlyFire":                  "TrueFalse", //bEnableFriendlyFire=False,
		"bEnableInvaderEnemy":                  "TrueFalse", //bEnableInvaderEnemy=True,
		"bActiveUNKO":                          "TrueFalse", //bActiveUNKO=False,
		"bEnableAimAssistPad":                  "TrueFalse", //bEnableAimAssistPad=True,
		"bEnableAimAssistKeyboard":             "TrueFalse", //bEnableAimAssistKeyboard=False,
		"DropItemMaxNum":                       "Numeric",   //DropItemMaxNum=3000,
		"DropItemMaxNum_UNKO":                  "Numeric",   //DropItemMaxNum_UNKO=100,
		"BaseCampMaxNum":                       "Numeric",   //BaseCampMaxNum=128,
		"BaseCampWorkerMaxNum":                 "Numeric",   //BaseCampWorkerMaxNum=15,
		"DropItemAliveMaxHours":                "Floating",  //DropItemAliveMaxHours=1.000000,
		"AutoResetGuildTimeNoOnlinePlayers":    "Floating",  //AutoResetGuildTimeNoOnlinePlayers=72.000000,
		"bAutoResetGuildNoOnlinePlayers":       "TrueFalse", //bAutoResetGuildNoOnlinePlayers=False,
		"GuildPlayerMaxNum":                    "Numeric",   //GuildPlayerMaxNum=20,
		"PalEggDefaultHatchingTime":            "Floating",  //PalEggDefaultHatchingTime=72.000000,
		"WorkSpeedRate":                        "Floating",  //WorkSpeedRate=1.000000,
		"bIsMultiplay":                         "TrueFalse", //bIsMultiplay=False,
		"bIsPvP":                               "TrueFalse", //bIsPvP=False,
		"bCanPickupOtherGuildDeathPenaltyDrop": "TrueFalse", //bCanPickupOtherGuildDeathPenaltyDrop=False,
		"bEnableNonLoginPenalty":               "TrueFalse", //bEnableNonLoginPenalty=True,
		"bEnableFastTravel":                    "TrueFalse", //bEnableFastTravel=True,
		"bIsStartLocationSelectByMap":          "TrueFalse", //bIsStartLocationSelectByMap=True,
		"bExistPlayerAfterLogout":              "TrueFalse", //bExistPlayerAfterLogout=False,
		"bEnableDefenseOtherGuildPlayer":       "TrueFalse", //bEnableDefenseOtherGuildPlayer=False,
		"CoopPlayerMaxNum":                     "Numeric",   //CoopPlayerMaxNum=4,
		"ServerPlayerMaxNum":                   "Numeric",   //ServerPlayerMaxNum=32,
		"ServerName":                           "String",    //ServerName="Default Palworld Server",
		"ServerDescription":                    "String",    //ServerDescription="",
		"ServerPassword":                       "AlphaDash", //ServerPassword="",
		"AdminPassword":                        "AlphaDash", //AdminPassword="",
		"PublicIP":                             "String",    //PublicIP="",
		"PublicPort":                           "Numeric",   //PublicPort=8211,
		"RCONPort":                             "Numeric",   //RCONPort=25575,
		"RCONEnabled":                          "TrueFalse", //RCONEnabled=False,
		"bUseAuth":                             "TrueFalse", //bUseAuth=True,
		"BanListURL":                           "String",    //BanListURL="https://api.palworldgame.com/api/banlist.txt"
		"Region":                               "String",    //Region="",
		"bShowPlayerList":                      "TrueFalse", //bShowPlayerList=False
		"RESTAPIEnabled":                       "TrueFalse", //RESTAPIEnabled=False
		"RESTAPIPort":                          "Numeric",   //RESTAPIPort=8212
		"AllowConnectPlatform":                 "String",    //AllowConnectPlatform=Steam
		"bIsUseBackupSaveData":                 "TrueFalse", //bIsUseBackupSaveData=True
		"LogFormatType":                        "String",    //LogFormatType=Text
		// Add other keys as needed
	}

	// Specify keys for which quotes should be added
	envVarsQuotes := map[string]bool{
		"ServerName":        true,
		"ServerPassword":    true,
		"AdminPassword":     true,
		"ServerDescription": true,
		"BanListURL":        true,
		"PublicIP":          true,
		// Add other keys as needed
	}
	// Determine the operating system
	var osFolder string
	switch runtime.GOOS {
	case "windows":
		osFolder = "WindowsServer"
	case "linux":
		// Check if the WINEPREFIX environment variable exists
		if _, winePrefixExists := os.LookupEnv("WINEPREFIX"); winePrefixExists {
			osFolder = "WindowsServer"
		} else if _, err := exec.LookPath("proton"); err == nil {
			osFolder = "WindowsServer"
		} else {
			osFolder = "LinuxServer"
		}
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
	} else if fileInfo.Size() == 0 || fileInfo.Size() < 1200 {
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
	iniOpenFIle, err := os.Open(iniFilePath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}

	defer iniOpenFIle.Close()

	iniContent, err := io.ReadAll(iniOpenFIle)
	if err != nil {
		fmt.Printf("Error reading INI file: %v\n", err)
		return
	}

	// Update values based on environment variables
	for key, value := range envVars {

		//val is the value that is in the enviroment variable, ok is true or false based of it exitis
		//LookupEnv retrieves the value of the environment variable named by the key. If the variable is present in the environment the value (which may be empty) is returned and the boolean is true. Otherwise the returned value will be empty and the boolean will be false.
		val, ok := os.LookupEnv(value)

		// If the environment variable doesn't exist skip
		if !ok {
			continue
		}

		//The variable exitis but is empty, so it will fail the validation so just set it to empty
		if ok && val == "" {
			fmt.Printf("Updating empty key: %s\n", key)
			setINIValue(&iniContent, key, "", envVarsQuotes[key])
		}

		//The Variable is set and it is not empty so validate and if true then set it in the file
		if ok && val != "" {
			// Check if there's a validation rule for the key
			if ruleName, ok := envVarsValidationRules[key]; ok {
				// Check if there's a validation function for the rule name
				if rule, ok := ValidationRules[ruleName]; ok {
					// Validate the value based on the rule
					if !rule(val) {
						fmt.Printf("Validation failed for key: %s, value: %s\n", key, val)
						continue
					}
				} else {
					fmt.Printf("No validation rule found for key: %s\n", key)
				}
			} else {
				fmt.Printf("No validation rule specified for key: %s\n", key)
			}

			// Update the value in the INI file
			fmt.Printf("Updating key: %s with value: %s\n", key, val)
			setINIValue(&iniContent, key, val, envVarsQuotes[key])

		}
	}

	// Write the updated contents back to the INI file
	err = os.WriteFile(iniFilePath, iniContent, 0644)
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

	// Line 340-364 is chatgpt generated but seems to work just fine
	var endPos int
	// normal case, key=value,
	endPos_1 := strings.Index(contentStr[pos:], ",")
	// Edge case as the last key has no ending ,
	endPos_2 := strings.Index(contentStr[pos:], ")")

	// Check if both endPos_1 and endPos_2 are -1, indicating neither comma nor closing parenthesis was found
	if endPos_1 == -1 && endPos_2 == -1 {
		// Set endPos to the end of the string
		endPos = len(contentStr)
	} else {
		// If either endPos_1 or endPos_2 is -1, replace it with a large value
		if endPos_1 == -1 {
			endPos_1 = len(contentStr) + 1
		}
		if endPos_2 == -1 {
			endPos_2 = len(contentStr) + 1
		}

		// Choose the minimum of endPos_1 and endPos_2
		if endPos_1 <= endPos_2 {
			endPos = endPos_1
		} else {
			endPos = endPos_2
		}
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
	srcOpenFile, err := os.Open(src)

	if err != nil {
		return err
	}
	defer srcOpenFile.Close()

	data, err := io.ReadAll(srcOpenFile)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}

func getIPAddressKey() string {
	// Check if PUBLIC_IP environment variable exists and is not empty
	val, ok := os.LookupEnv("PUBLIC_IP")
	if ok && val != "" {
		return "PUBLIC_IP"
	}

	// Fallback to SERVER_IP if PUBLIC_IP is empty or does not exist
	return "SERVER_IP"
}
