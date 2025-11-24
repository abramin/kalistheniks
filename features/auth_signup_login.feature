Feature: User signup and login to receive a JWT token
  As a new user
  I want to sign up and log in
  So that I can receive a token to access protected endpoints

  Background:
    Given the database is empty

  Scenario: Successful signup and login returns a valid token
    When I POST /signup with body:
      """
      {"email":"user@example.com","password":"StrongPass!1"}
      """
    Then the response status should be 201
    And the response JSON should include "user.id" and "token"
    When I POST /login with body:
      """
      {"email":"user@example.com","password":"StrongPass!1"}
      """
    Then the response status should be 200
    And the response JSON should include a non-empty "token"

  Scenario: Duplicate signup is rejected
    Given a user already exists with email "user@example.com"
    When I POST /signup with body:
      """
      {"email":"user@example.com","password":"AnotherPass!2"}
      """
    Then the response status should be 400
    And the response JSON should include an "error" explaining the email is taken

  Scenario: Login fails with wrong password
    Given a user exists with email "user@example.com" and password "StrongPass!1"
    When I POST /login with body:
      """
      {"email":"user@example.com","password":"WrongPass"}
      """
    Then the response status should be 401
    And the response JSON should include "error":"invalid credentials"

  Scenario: Signup fails with missing fields
    When I POST /signup with body:
      """
      {"email":"","password":""}
      """
    Then the response status should be 400
    And the response JSON should include an "error" about invalid request body
