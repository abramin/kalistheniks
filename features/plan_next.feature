Feature: Request next workout plan suggestion
  As an authenticated user
  I want to get the next progression suggestion
  So I can follow the program

  Background:
    Given I have a valid token from logging in as "user@example.com"

  Scenario: Suggest increased weight after hitting upper rep range
    Given my last recorded set has:
      | exercise_id | squat-uuid |
      | reps        | 12         |
      | weight_kg   | 80.0       |
      | session_type| lower      |
    When I GET /plan/next with headers:
      | Authorization | Bearer <token> |
    Then the response status should be 200
    And the response JSON should include:
      | exercise_id | squat-uuid          |
      | weight_kg   | 82.5                |
      | reps        | 12                  |
      | notes       | contains "increase" |

  Scenario: Suggest reduced reps after early failure
    Given my last recorded set has:
      | exercise_id | bench-uuid |
      | reps        | 5          |
      | weight_kg   | 70.0       |
      | session_type| upper      |
    When I GET /plan/next with headers:
      | Authorization | Bearer <token> |
    Then the response status should be 200
    And the response JSON should include:
      | exercise_id | bench-uuid              |
      | weight_kg   | 70.0                    |
      | reps        | 5 or less               |
      | notes       | contains "reduce reps"  |

  Scenario: Suggest default when no history exists
    Given I have no recorded sessions or sets
    When I GET /plan/next with headers:
      | Authorization | Bearer <token> |
    Then the response status should be 200
    And the response JSON should include default values:
      | weight_kg | 20          |
      | reps      | 8           |
      | notes     | contains "No history" |

  Scenario: Plan request fails with missing token
    When I GET /plan/next without an Authorization header
    Then the response status should be 401
    And the response JSON should include "error":"missing token"
