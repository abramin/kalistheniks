Feature: Create a training session
  As an authenticated user
  I want to create a training session
  So I can track my workouts

  Background:
    Given I have a valid token from logging in as "user@example.com"

  Scenario: Successful session creation
    When I POST /sessions with headers:
      | Authorization | Bearer <token> |
    And body:
      """
      {"performed_at":"2024-01-01T10:00:00Z","session_type":"upper","notes":"Push day"}
      """
    Then the response status should be 201
    And the response JSON should include "id","user_id","performed_at","session_type","notes"

  Scenario: Session creation fails with missing token
    When I POST /sessions without an Authorization header
    Then the response status should be 401
    And the response JSON should include "error":"missing token"

  Scenario: Session creation fails with invalid body
    When I POST /sessions with headers:
      | Authorization | Bearer <token> |
    And body:
      """
      {"performed_at":"not-a-timestamp"}
      """
    Then the response status should be 400
    And the response JSON should include "error":"invalid request body"
