package server

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/solo7.media/platform.drifter.solo7.media/internal/domain"
	"github.com/solo7.media/platform.drifter.solo7.media/internal/server"
)

// MockLogger implements the Logger interface for testing
type MockLogger struct{}

func (m *MockLogger) Debug(msg string, fields map[string]interface{}) {}
func (m *MockLogger) Info(msg string, fields map[string]interface{})  {}
func (m *MockLogger) Warn(msg string, fields map[string]interface{})  {}
func (m *MockLogger) Error(msg string, fields map[string]interface{}) {}
func (m *MockLogger) Fatal(msg string, fields map[string]interface{}) {}

func TestRulesEngine_RegisterAndUnregisterRule(t *testing.T) {
	logger := &MockLogger{}
	engine := server.NewRulesEngine(logger)
	ctx := context.Background()

	// Test registering a rule
	rule := &domain.GameRule{
		ID:       uuid.New(),
		Triggers: []domain.EventTrigger{{Type: "combat_start"}},
		Actions:  []domain.Action{{Type: "notification", Target: "combat", Properties: map[string]interface{}{"message": "Combat started!"}}},
		Priority: 10,
	}

	err := engine.RegisterRule(ctx, rule)
	if err != nil {
		t.Fatalf("Failed to register rule: %v", err)
	}

	// Test unregistering the rule
	err = engine.UnregisterRule(ctx, rule.ID)
	if err != nil {
		t.Fatalf("Failed to unregister rule: %v", err)
	}

	// Test unregistering non-existent rule
	err = engine.UnregisterRule(ctx, uuid.New())
	if err == nil {
		t.Fatal("Expected error when unregistering non-existent rule")
	}
}

func TestRulesEngine_ProcessEvent_SimpleConditions(t *testing.T) {
	logger := &MockLogger{}
	engine := server.NewRulesEngine(logger)
	ctx := context.Background()

	// Register a rule with simple conditions
	rule := &domain.GameRule{
		ID:       uuid.New(),
		Triggers: []domain.EventTrigger{{Type: "player_action"}},
		Conditions: []domain.Condition{
			{Type: "simple", Property: "action_type", Operator: "==", Value: "attack"},
		},
		Actions: []domain.Action{
			{
				Type:   "notification",
				Target: "combat",
				Properties: map[string]interface{}{
					"message": "Player attacks!",
				},
			},
		},
		Priority: 10,
	}

	err := engine.RegisterRule(ctx, rule)
	if err != nil {
		t.Fatalf("Failed to register rule: %v", err)
	}

	// Test event that should trigger the rule
	eventData := map[string]interface{}{
		"action_type": "attack",
		"player_id":   uuid.New().String(),
		"self": map[string]interface{}{
			"level": 5.0,
		},
	}

	result, err := engine.ProcessEvent(ctx, "player_action", eventData)
	if err != nil {
		t.Fatalf("Failed to process event: %v", err)
	}

	if len(result.ClientNotifications) != 1 {
		t.Fatalf("Expected 1 notification, got %d", len(result.ClientNotifications))
	}

	// Test event that should not trigger the rule
	eventData["action_type"] = "defend"
	result, err = engine.ProcessEvent(ctx, "player_action", eventData)
	if err != nil {
		t.Fatalf("Failed to process event: %v", err)
	}

	if len(result.ClientNotifications) != 0 {
		t.Fatalf("Expected 0 notifications, got %d", len(result.ClientNotifications))
	}
}

func TestRulesEngine_ProcessEvent_ExpressionConditions(t *testing.T) {
	logger := &MockLogger{}
	engine := server.NewRulesEngine(logger)
	ctx := context.Background()

	// Register a rule with expression conditions
	rule := &domain.GameRule{
		ID:       uuid.New(),
		Triggers: []domain.EventTrigger{{Type: "combat_damage"}},
		Conditions: []domain.Condition{
			{Type: "expression", Property: "damage > 10 && target.armor < 15"},
		},
		Actions: []domain.Action{
			{
				Type:   "notification",
				Target: "combat",
				Properties: map[string]interface{}{
					"message": "Critical hit!",
				},
			},
		},
		Priority: 10,
	}

	err := engine.RegisterRule(ctx, rule)
	if err != nil {
		t.Fatalf("Failed to register rule: %v", err)
	}

	// Test event that should trigger the rule
	eventData := map[string]interface{}{
		"damage": 15.0,
		"target": map[string]interface{}{
			"armor": 10.0,
		},
		"player_id": uuid.New().String(),
	}

	result, err := engine.ProcessEvent(ctx, "combat_damage", eventData)
	if err != nil {
		t.Fatalf("Failed to process event: %v", err)
	}

	if len(result.ClientNotifications) != 1 {
		t.Fatalf("Expected 1 notification, got %d", len(result.ClientNotifications))
	}

	// Test event that should not trigger the rule
	eventData["damage"] = 5.0
	result, err = engine.ProcessEvent(ctx, "combat_damage", eventData)
	if err != nil {
		t.Fatalf("Failed to process event: %v", err)
	}

	if len(result.ClientNotifications) != 0 {
		t.Fatalf("Expected 0 notifications, got %d", len(result.ClientNotifications))
	}
}

func TestRulesEngine_ProcessEvent_ExpressionActions(t *testing.T) {
	logger := &MockLogger{}
	engine := server.NewRulesEngine(logger)
	ctx := context.Background()

	// Register a rule with expression actions
	rule := &domain.GameRule{
		ID:       uuid.New(),
		Triggers: []domain.EventTrigger{{Type: "player_heal"}},
		Actions: []domain.Action{
			{
				Type:   "expression",
				Target: "heal_calculation",
				Properties: map[string]interface{}{
					"expression": "roll(\"2d6\") + self.level",
				},
			},
			{
				Type:   "notification",
				Target: "healing",
				Properties: map[string]interface{}{
					"message": "Healed for {{heal_amount}} points!",
				},
			},
		},
		Priority: 10,
	}

	err := engine.RegisterRule(ctx, rule)
	if err != nil {
		t.Fatalf("Failed to register rule: %v", err)
	}

	// Test event
	eventData := map[string]interface{}{
		"player_id": uuid.New().String(),
		"self": map[string]interface{}{
			"level": 5.0,
		},
	}

	result, err := engine.ProcessEvent(ctx, "player_heal", eventData)
	if err != nil {
		t.Fatalf("Failed to process event: %v", err)
	}

	// Should have executed both actions
	if len(result.ClientNotifications) != 1 {
		t.Fatalf("Expected 1 notification, got %d", len(result.ClientNotifications))
	}
}

func TestRulesEngine_ProcessEvent_StateChanges(t *testing.T) {
	logger := &MockLogger{}
	engine := server.NewRulesEngine(logger)
	ctx := context.Background()

	entityID := uuid.New()

	// Register a rule that modifies entity state
	rule := &domain.GameRule{
		ID:       uuid.New(),
		Triggers: []domain.EventTrigger{{Type: "level_up"}},
		Actions: []domain.Action{
			{
				Type:   "state_change",
				Target: "gameplay",
				Properties: map[string]interface{}{
					"level":      "self.level + 1",
					"health":     "self.max_health + roll(\"1d8\")",
					"experience": 0,
				},
			},
		},
		Priority: 10,
	}

	err := engine.RegisterRule(ctx, rule)
	if err != nil {
		t.Fatalf("Failed to register rule: %v", err)
	}

	// Test event
	eventData := map[string]interface{}{
		"entity_id": entityID.String(),
		"self": map[string]interface{}{
			"level":      5.0,
			"max_health": 50.0,
		},
	}

	result, err := engine.ProcessEvent(ctx, "level_up", eventData)
	if err != nil {
		t.Fatalf("Failed to process event: %v", err)
	}

	if len(result.WorldStateChanges) != 1 {
		t.Fatalf("Expected 1 state change, got %d", len(result.WorldStateChanges))
	}

	change := result.WorldStateChanges[0]
	if change.EntityID != entityID {
		t.Fatalf("Expected entity ID %s, got %s", entityID, change.EntityID)
	}

	if change.Component != "gameplay" {
		t.Fatalf("Expected component 'gameplay', got '%s'", change.Component)
	}

	// Check that expressions were evaluated
	if level, exists := change.Changes["level"]; !exists {
		t.Fatal("Expected 'level' in changes")
	} else if level != 6.0 {
		t.Fatalf("Expected level 6.0, got %v", level)
	}
}

func TestRulesEngine_ProcessEvent_AestheticEvents(t *testing.T) {
	logger := &MockLogger{}
	engine := server.NewRulesEngine(logger)
	ctx := context.Background()

	entityID := uuid.New()

	// Register a rule that triggers aesthetic events
	rule := &domain.GameRule{
		ID:       uuid.New(),
		Triggers: []domain.EventTrigger{{Type: "spell_cast"}},
		Actions: []domain.Action{
			{
				Type:   "aesthetic_event",
				Target: "particle_effect",
				Properties: map[string]interface{}{
					"effect_type": "fireball",
					"intensity":   "self.level * 2",
					"duration":    "3.0",
				},
			},
		},
		Priority: 10,
	}

	err := engine.RegisterRule(ctx, rule)
	if err != nil {
		t.Fatalf("Failed to register rule: %v", err)
	}

	// Test event
	eventData := map[string]interface{}{
		"entity_id": entityID.String(),
		"self": map[string]interface{}{
			"level": 5.0,
		},
	}

	result, err := engine.ProcessEvent(ctx, "spell_cast", eventData)
	if err != nil {
		t.Fatalf("Failed to process event: %v", err)
	}

	if len(result.AestheticEvents) != 1 {
		t.Fatalf("Expected 1 aesthetic event, got %d", len(result.AestheticEvents))
	}

	event := result.AestheticEvents[0]
	if event.EntityID != entityID {
		t.Fatalf("Expected entity ID %s, got %s", entityID, event.EntityID)
	}

	if event.Type != "particle_effect" {
		t.Fatalf("Expected type 'particle_effect', got '%s'", event.Type)
	}

	// Check that expressions were evaluated
	if intensity, exists := event.Properties["intensity"]; !exists {
		t.Fatal("Expected 'intensity' in properties")
	} else if intensity != 10.0 {
		t.Fatalf("Expected intensity 10.0, got %v", intensity)
	}
}

func TestRulesEngine_ProcessEvent_RulePriority(t *testing.T) {
	logger := &MockLogger{}
	engine := server.NewRulesEngine(logger)
	ctx := context.Background()

	// Register multiple rules with different priorities
	rule1 := &domain.GameRule{
		ID:       uuid.New(),
		Triggers: []domain.EventTrigger{{Type: "combat_damage"}},
		Actions: []domain.Action{
			{
				Type:   "notification",
				Target: "combat",
				Properties: map[string]interface{}{
					"message": "Low priority rule",
				},
			},
		},
		Priority: 1,
	}

	rule2 := &domain.GameRule{
		ID:       uuid.New(),
		Triggers: []domain.EventTrigger{{Type: "combat_damage"}},
		Actions: []domain.Action{
			{
				Type:   "notification",
				Target: "combat",
				Properties: map[string]interface{}{
					"message": "High priority rule",
				},
			},
		},
		Priority: 10,
	}

	err := engine.RegisterRule(ctx, rule1)
	if err != nil {
		t.Fatalf("Failed to register rule1: %v", err)
	}

	err = engine.RegisterRule(ctx, rule2)
	if err != nil {
		t.Fatalf("Failed to register rule2: %v", err)
	}

	// Test event
	eventData := map[string]interface{}{
		"damage":    10.0,
		"player_id": uuid.New().String(),
	}

	result, err := engine.ProcessEvent(ctx, "combat_damage", eventData)
	if err != nil {
		t.Fatalf("Failed to process event: %v", err)
	}

	// Should have 2 notifications (both rules triggered)
	if len(result.ClientNotifications) != 2 {
		t.Fatalf("Expected 2 notifications, got %d", len(result.ClientNotifications))
	}

	// High priority rule should be executed first
	if result.ClientNotifications[0].Message != "High priority rule" {
		t.Fatalf("Expected high priority rule first, got: %s", result.ClientNotifications[0].Message)
	}

	if result.ClientNotifications[1].Message != "Low priority rule" {
		t.Fatalf("Expected low priority rule second, got: %s", result.ClientNotifications[1].Message)
	}
}

func TestRulesEngine_ProcessEvent_ComplexExpressions(t *testing.T) {
	logger := &MockLogger{}
	engine := server.NewRulesEngine(logger)
	ctx := context.Background()

	// Register a rule with complex expressions
	rule := &domain.GameRule{
		ID:       uuid.New(),
		Triggers: []domain.EventTrigger{{Type: "spell_damage"}},
		Conditions: []domain.Condition{
			{Type: "expression", Property: "spell_level >= 3 && target.resistance < 10"},
		},
		Actions: []domain.Action{
			{
				Type:   "state_change",
				Target: "gameplay",
				Properties: map[string]interface{}{
					"damage": "roll(\"8d6\") + self.spell_power + (spell_level * 2)",
				},
			},
		},
		Priority: 10,
	}

	err := engine.RegisterRule(ctx, rule)
	if err != nil {
		t.Fatalf("Failed to register rule: %v", err)
	}

	// Test event that should trigger
	eventData := map[string]interface{}{
		"spell_level": 5.0,
		"entity_id":   uuid.New().String(),
		"self": map[string]interface{}{
			"spell_power": 8.0,
		},
		"target": map[string]interface{}{
			"resistance": 5.0,
		},
	}

	result, err := engine.ProcessEvent(ctx, "spell_damage", eventData)
	if err != nil {
		t.Fatalf("Failed to process event: %v", err)
	}

	if len(result.WorldStateChanges) != 1 {
		t.Fatalf("Expected 1 state change, got %d", len(result.WorldStateChanges))
	}

	// Test event that should not trigger
	eventData["target"].(map[string]interface{})["resistance"] = 15.0
	result, err = engine.ProcessEvent(ctx, "spell_damage", eventData)
	if err != nil {
		t.Fatalf("Failed to process event: %v", err)
	}

	if len(result.WorldStateChanges) != 0 {
		t.Fatalf("Expected 0 state changes, got %d", len(result.WorldStateChanges))
	}
}

func TestRulesEngine_GetActiveRules(t *testing.T) {
	logger := &MockLogger{}
	engine := server.NewRulesEngine(logger)
	ctx := context.Background()

	// Register some rules
	rule1 := &domain.GameRule{
		ID:       uuid.New(),
		Triggers: []domain.EventTrigger{{Type: "combat_start"}},
		Priority: 5,
	}

	rule2 := &domain.GameRule{
		ID:       uuid.New(),
		Triggers: []domain.EventTrigger{{Type: "player_action"}},
		Priority: 10,
	}

	err := engine.RegisterRule(ctx, rule1)
	if err != nil {
		t.Fatalf("Failed to register rule1: %v", err)
	}

	err = engine.RegisterRule(ctx, rule2)
	if err != nil {
		t.Fatalf("Failed to register rule2: %v", err)
	}

	// Get active rules
	regionID := uuid.New()
	rules, err := engine.GetActiveRules(ctx, regionID)
	if err != nil {
		t.Fatalf("Failed to get active rules: %v", err)
	}

	if len(rules) != 2 {
		t.Fatalf("Expected 2 active rules, got %d", len(rules))
	}
}

func TestRulesEngine_ErrorHandling(t *testing.T) {
	logger := &MockLogger{}
	engine := server.NewRulesEngine(logger)
	ctx := context.Background()

	// Test processing event with invalid data
	eventData := map[string]interface{}{
		"invalid_field": "test",
	}

	result, err := engine.ProcessEvent(ctx, "test_event", eventData)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should handle gracefully without crashing
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
}

func TestRulesEngine_ConcurrentAccess(t *testing.T) {
	logger := &MockLogger{}
	engine := server.NewRulesEngine(logger)
	ctx := context.Background()

	// Test concurrent rule registration and event processing
	done := make(chan bool, 10)

	// Concurrent rule registration
	for i := 0; i < 5; i++ {
		go func(i int) {
			rule := &domain.GameRule{
				ID:       uuid.New(),
				Triggers: []domain.EventTrigger{{Type: "test_event"}},
				Priority: i,
			}
			engine.RegisterRule(ctx, rule)
			done <- true
		}(i)
	}

	// Concurrent event processing
	for i := 0; i < 5; i++ {
		go func() {
			eventData := map[string]interface{}{
				"test": "value",
			}
			engine.ProcessEvent(ctx, "test_event", eventData)
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should not have crashed
	t.Log("Concurrent access test completed successfully")
}
