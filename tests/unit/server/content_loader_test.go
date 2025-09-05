package server

import (
	"fmt"
	"testing"
	"time"

	"github.com/solo-seven/platform.drifter.solo7.media/internal/logger"
	"github.com/solo-seven/platform.drifter.solo7.media/internal/server"
)

// Test config implementation
type testConfig struct{}

func (c testConfig) GetServerPort() int                  { return 8080 }
func (c testConfig) GetMaxConnections() int              { return 100 }
func (c testConfig) GetHeartbeatInterval() time.Duration { return 30 * time.Second }
func (c testConfig) GetRegionSize() float64              { return 1000.0 }
func (c testConfig) GetMaxEntitiesPerRegion() int        { return 1000 }
func (c testConfig) GetLogLevel() string                 { return "INFO" }
func (c testConfig) GetDatabaseURL() string              { return "" }
func (c testConfig) GetRedisURL() string                 { return "" }
func (c testConfig) GetLogVerbose() bool                 { return false }
func (c testConfig) GetLogOutput() string                { return "console" }
func (c testConfig) GetLogFilePath() string              { return "" }
func (c testConfig) GetLogServiceName() string           { return "test-server" }

func TestEquipmentLoading(t *testing.T) {
	// Create logger
	loggerFactory := logger.NewLoggerFactory()
	gameLogger, err := loggerFactory.CreateLoggerFromConfig(testConfig{}, false)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Create content repository manager and TOML parser
	contentRepoManager := server.NewContentRepositoryManager(gameLogger)
	tomlParser := server.NewTOMLParser(gameLogger)

	// Create content loader
	contentLoader := server.NewContentLoader(
		gameLogger,
		tomlParser,
		contentRepoManager,
		"../../../examples/shattered-realms",
	)

	// Load all content
	if err := contentLoader.LoadAllContent(); err != nil {
		t.Fatalf("Failed to load content: %v", err)
	}

	// Test that items were loaded
	allItems := contentLoader.GetAllItems()
	if len(allItems) == 0 {
		t.Error("Expected items to be loaded, but got 0 items")
		return
	}

	// Find the specific items from equipment.toml
	var stabilizerCrystal *server.ItemData
	var chaosBlade *server.ItemData

	for _, item := range allItems {
		switch item.ID {
		case "item.stabilizer_crystal":
			stabilizerCrystal = item
		case "item.chaos_blade":
			chaosBlade = item
		}
	}

	// Test Stabilizer Crystal
	if stabilizerCrystal == nil {
		t.Error("Stabilizer Crystal not found in loaded items")
	} else {
		t.Run("StabilizerCrystal", func(t *testing.T) {
			if stabilizerCrystal.Name != "Stabilizer Crystal" {
				t.Errorf("Expected name 'Stabilizer Crystal', got '%s'", stabilizerCrystal.Name)
			}
			if stabilizerCrystal.Type != "wondrous_item" {
				t.Errorf("Expected type 'wondrous_item', got '%s'", stabilizerCrystal.Type)
			}

			// Test additional properties stored in Properties map
			if rarity, exists := stabilizerCrystal.Properties["rarity"]; exists {
				if rarity != "uncommon" {
					t.Errorf("Expected rarity 'uncommon', got '%s'", rarity)
				}
			} else {
				t.Error("rarity property not found")
			}

			if attunementRequired, exists := stabilizerCrystal.Properties["attunement_required"]; exists {
				if attunementRequired != false {
					t.Errorf("Expected attunement_required false, got %v", attunementRequired)
				}
			} else {
				t.Error("attunement_required property not found")
			}

			// Test properties
			if stabilizerCrystal.Properties == nil {
				t.Error("Expected properties to be loaded")
			} else {
				// Test reality_anchor property
				if realityAnchor, exists := stabilizerCrystal.Properties["reality_anchor"]; exists {
					if realityAnchorMap, ok := realityAnchor.(map[string]interface{}); ok {
						if effect, exists := realityAnchorMap["effect"]; exists {
							expectedEffect := "increase_local_stability(amount=10, radius=30)"
							if effect != expectedEffect {
								t.Errorf("Expected reality_anchor effect '%s', got '%s'", expectedEffect, effect)
							}
						} else {
							t.Error("reality_anchor effect not found")
						}
						if duration, exists := realityAnchorMap["duration"]; exists {
							expectedDuration := "permanent_while_held"
							if duration != expectedDuration {
								t.Errorf("Expected reality_anchor duration '%s', got '%s'", expectedDuration, duration)
							}
						} else {
							t.Error("reality_anchor duration not found")
						}
					} else {
						t.Error("reality_anchor property is not a map")
					}
				} else {
					t.Error("reality_anchor property not found")
				}

				// Test fragile property
				if fragile, exists := stabilizerCrystal.Properties["fragile"]; exists {
					if fragileMap, ok := fragile.(map[string]interface{}); ok {
						if condition, exists := fragileMap["condition"]; exists {
							expectedCondition := "reality_stability <= 15"
							if condition != expectedCondition {
								t.Errorf("Expected fragile condition '%s', got '%s'", expectedCondition, condition)
							}
						} else {
							t.Error("fragile condition not found")
						}
						if effect, exists := fragileMap["effect"]; exists {
							expectedEffect := "item_destroyed(); reality_stability -= 5"
							if effect != expectedEffect {
								t.Errorf("Expected fragile effect '%s', got '%s'", expectedEffect, effect)
							}
						} else {
							t.Error("fragile effect not found")
						}
					} else {
						t.Error("fragile property is not a map")
					}
				} else {
					t.Error("fragile property not found")
				}
			}
		})
	}

	// Test Chaos Blade
	if chaosBlade == nil {
		t.Error("Chaos Blade not found in loaded items")
	} else {
		t.Run("ChaosBlade", func(t *testing.T) {
			if chaosBlade.Name != "Chaos Blade" {
				t.Errorf("Expected name 'Chaos Blade', got '%s'", chaosBlade.Name)
			}
			if chaosBlade.Type != "weapon" {
				t.Errorf("Expected type 'weapon', got '%s'", chaosBlade.Type)
			}

			// Test additional properties stored in Properties map
			if weaponType, exists := chaosBlade.Properties["weapon_type"]; exists {
				if weaponType != "longsword" {
					t.Errorf("Expected weapon_type 'longsword', got '%s'", weaponType)
				}
			} else {
				t.Error("weapon_type property not found")
			}

			if damage, exists := chaosBlade.Properties["damage"]; exists {
				if damage != "1d8" {
					t.Errorf("Expected damage '1d8', got '%s'", damage)
				}
			} else {
				t.Error("damage property not found")
			}

			if magical, exists := chaosBlade.Properties["magical"]; exists {
				if magical != true {
					t.Errorf("Expected magical true, got %v", magical)
				}
			} else {
				t.Error("magical property not found")
			}

			if attunementRequired, exists := chaosBlade.Properties["attunement_required"]; exists {
				if attunementRequired != true {
					t.Errorf("Expected attunement_required true, got %v", attunementRequired)
				}
			} else {
				t.Error("attunement_required property not found")
			}

			// Test properties
			if chaosBlade.Properties == nil {
				t.Error("Expected properties to be loaded")
			} else {
				// Test variable_damage property
				if variableDamage, exists := chaosBlade.Properties["variable_damage"]; exists {
					if variableDamageMap, ok := variableDamage.(map[string]interface{}); ok {
						if effect, exists := variableDamageMap["effect"]; exists {
							// Check that the effect contains the expected multiline string
							effectStr := fmt.Sprintf("%v", effect)
							expectedParts := []string{
								"base_damage = roll(\"1d8\")",
								"if (reality_stability < 50)",
								"extra_damage = roll(\"1d6\")",
								"damage_type = random_choice([\"fire\", \"cold\", \"lightning\", \"thunder\"])",
								"return base_damage + extra_damage",
							}
							for _, part := range expectedParts {
								if !contains(effectStr, part) {
									t.Errorf("Expected variable_damage effect to contain '%s', but it didn't", part)
								}
							}
						} else {
							t.Error("variable_damage effect not found")
						}
					} else {
						t.Error("variable_damage property is not a map")
					}
				} else {
					t.Error("variable_damage property not found")
				}

				// Test reality_disruption property
				if realityDisruption, exists := chaosBlade.Properties["reality_disruption"]; exists {
					if realityDisruptionMap, ok := realityDisruption.(map[string]interface{}); ok {
						if onCritical, exists := realityDisruptionMap["on_critical"]; exists {
							expectedOnCritical := "target_location.reality_stability -= roll(\"1d4\")"
							if onCritical != expectedOnCritical {
								t.Errorf("Expected reality_disruption on_critical '%s', got '%s'", expectedOnCritical, onCritical)
							}
						} else {
							t.Error("reality_disruption on_critical not found")
						}
					} else {
						t.Error("reality_disruption property is not a map")
					}
				} else {
					t.Error("reality_disruption property not found")
				}
			}
		})
	}

	// Test that we can get items by ID
	t.Run("GetItemByID", func(t *testing.T) {
		stabilizerCrystalByID := contentLoader.GetItem("item.stabilizer_crystal")
		if stabilizerCrystalByID == nil {
			t.Error("Failed to get Stabilizer Crystal by ID")
		} else if stabilizerCrystalByID.Name != "Stabilizer Crystal" {
			t.Errorf("Expected Stabilizer Crystal by ID, got '%s'", stabilizerCrystalByID.Name)
		}

		chaosBladeByID := contentLoader.GetItem("item.chaos_blade")
		if chaosBladeByID == nil {
			t.Error("Failed to get Chaos Blade by ID")
		} else if chaosBladeByID.Name != "Chaos Blade" {
			t.Errorf("Expected Chaos Blade by ID, got '%s'", chaosBladeByID.Name)
		}
	})
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
