# GitHub Copilot Authentication Flow

This document outlines the end-to-end authentication process used to obtain a valid token for making requests to the GitHub Copilot API. The flow consists of two main phases:

1.  **GitHub Device Authorization Flow:** Obtaining a standard GitHub OAuth token by having the user authorize the application.
2.  **Copilot Token Exchange:** Exchanging the GitHub OAuth token for a specific, short-lived Copilot API token.

---

## Phase 1: GitHub Device Authorization Flow

This phase follows the standard GitHub OAuth Device Flow to get a user-authorized API token.

### Step 1.1: Request Device and User Codes

The application initiates the flow by making a `POST` request to GitHub.

*   **Endpoint:** `POST https://github.com/login/device/code`
*   **Headers:**
    *   `Accept: application/json`
*   **Body:**
    ```json
    {
      "client_id": "<GITHUB_CLIENT_ID>",
      "scope": "<REQUESTED_SCOPES>"
    }
    ```

GitHub responds with a device code, a user code for verification, and a polling interval.

*   **Example Response:**
    ```json
    {
      "device_code": "...",
      "user_code": "...",
      "verification_uri": "https://github.com/login/device",
      "expires_in": 900,
      "interval": 5
    }
    ```

### Step 1.2: Prompt User for Authorization

The application shows the `user_code` to the user and instructs them to visit the `verification_uri` to authorize the device.

### Step 1.3: Poll for Access Token

While waiting for the user to authorize, the application begins polling the token endpoint at the specified `interval`.

*   **Endpoint:** `POST https://github.com/login/oauth/access_token`
*   **Headers:**
    *   `Accept: application/json`
*   **Body:**
    ```json
    {
      "client_id": "<GITHUB_CLIENT_ID>",
      "device_code": "<DEVICE_CODE_FROM_STEP_1>",
      "grant_type": "urn:ietf:params:oauth:grant-type:device_code"
    }
    ```

Once the user completes authorization in the browser, the polling request will succeed and GitHub will respond with the user's OAuth token.

*   **Success Response:**
    ```json
    {
      "access_token": "gho_...",
      "token_type": "bearer",
      "scope": "..."
    }
    ```

This `access_token` is the **GitHub OAuth Token**.

---

## Phase 2: Copilot Token Exchange

With a valid GitHub OAuth token, the application can now exchange it for a token that is valid for the Copilot API.

### Step 2.1: Request Copilot Token

The application makes an authenticated `GET` request to a private Copilot endpoint.

*   **Endpoint:** `GET https://api.github.com/copilot_internal/v2/token`
*   **Headers:**
    *   `Authorization: Bearer <GITHUB_OAUTH_TOKEN_FROM_PHASE_1>`
    *   `Accept: application/json`

The response contains the final, short-lived Copilot token.

*   **Success Response:**
    ```json
    {
      "token": "tid=...;exp=...;...",
      "expires_at": 1672531200,
      "refresh_in": 1500
    }
    ```

This `token` is the one used in the `Authorization` header for all subsequent requests to `api.individual.githubcopilot.com`.

### Step 2.2: Automatic Token Refresh

The Copilot token is short-lived (e.g., expires in 30 minutes). The application is responsible for refreshing it automatically before it expires.

*   **Logic:** Use `setInterval` or a similar timer mechanism.
*   **Interval:** The refresh should be triggered after `(refresh_in - 60)` seconds to provide a buffer.
*   **Action:** The timer re-runs **Step 2.1** to get a new Copilot token and updates the application's state.
