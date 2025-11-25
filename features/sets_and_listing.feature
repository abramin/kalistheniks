Feature: Add sets to a session and list sessions with sets
  As an authenticated user
  I want to log sets for a session and see them in my history

  Background:
    Given I have a valid token from logging in as "user@example.com"
    And I have created a session with id "<session_id>"

  Scenario: Successfully add a set to an existing session
    When I POST /sessions/<session_id>/sets with headers:
      | Authorization | Bearer <token> |
    And body:
      """
      {"exercise_id":"deadlift-uuid","set_index":1,"reps":8,"weight_kg":100.0,"rpe":7}
      """
    Then the response status should be 201
    And the response JSON should include "id","session_id","exercise_id","set_index","reps","weight_kg","rpe"

  Scenario: Adding a set fails for an invalid session id
    When I POST /sessions/invalid-session-id/sets with headers:
      | Authorization | Bearer <token> |
    And body:
      """
      {"exercise_id":"deadlift-uuid","set_index":1,"reps":8,"weight_kg":100.0}
      """
    Then the response status should be 500
    And the response JSON field "error" should contain "failed"

  Scenario: Retrieve sessions with nested sets
    Given I have added two sets to session "<session_id>"
    When I GET /sessions with headers:
      | Authorization | Bearer <token> |
    Then the response status should be 200
    And the response JSON should include a list where:
      | field                   | expectation           |
      | [0].id                  | equals "<session_id>" |
      | [0].sets.length         | equals 2              |
      | [0].sets[0].exercise_id | equals "deadlift-uuid"|
      | [0].sets[0].reps        | equals 8              |
      | [0].sets[1].set_index   | equals 2              |

  Scenario: Listing sessions fails with invalid token
    When I GET /sessions with headers:
      | Authorization | Bearer invalid.token |
    Then the response status should be 401
    And the response JSON field "error" should be "invalid token"
