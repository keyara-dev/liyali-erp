# Profile Picture Persistence Fix

## Problem

User uploaded profile picture in settings page, but it didn't appear in navigation menus after upload or after login.

## Root Cause

The backend stored the avatar in the `preferences` JSONB field, but the `UserResponse` type used in login/register responses didn't include the `preferences` field. This meant:

1. When users logged in, the backend returned user data without preferences/avatar
2. The frontend session stored incomplete user data
3. Navigation components couldn't access the avatar because it wasn't in the session

## Solution

### Backend Changes

1. **Updated `UserResponse` type** (`backend/types/auth.go`)
   - Added `Preferences map[string]interface{}` field to include user preferences in API responses

2. **Updated `Login` method** (`backend/services/auth_service.go`)
   - Added code to unmarshal and include preferences in the login response
   - Ensures avatar is returned when user logs in

3. **Updated `Register` method** (`backend/services/auth_service.go`)
   - Added code to unmarshal and include preferences in the registration response
   - Ensures avatar is available for new users

### Frontend Changes

1. **Updated `loginAction`** (`frontend/src/app/_actions/auth.ts`)
   - Extracts avatar from `preferences.avatar`
   - Sets it at top-level `user.avatar` for easy component access
   - Ensures avatar is available immediately after login

2. **Updated `createNewAccount`** (`frontend/src/app/_actions/auth.ts`)
   - Extracts avatar from `preferences.avatar`
   - Sets it at top-level `user.avatar` for easy component access
   - Ensures avatar is available for new registrations

3. **Updated `updateAccountSettings`** (`frontend/src/app/_actions/settings.ts`)
   - Properly syncs all profile fields including avatar
   - Extracts avatar from preferences and sets at top level
   - Updates session immediately after settings change

4. **Updated `getUserProfile`** (`frontend/src/app/_actions/settings.ts`)
   - Extracts avatar from preferences to top-level for consistency
   - Ensures avatar is always available when profile is fetched

## How It Works Now

1. **Avatar Upload Flow:**
   - User uploads avatar in settings page
   - Backend stores it in `preferences.avatar` (JSONB field)
   - Frontend calls `updateAccountSettings` which updates the session
   - Session now has both `preferences.avatar` and top-level `avatar`
   - Navigation components immediately show the new avatar

2. **Login Flow:**
   - User logs in with credentials
   - Backend returns user data with `preferences` field
   - Frontend extracts `preferences.avatar` and sets it at `user.avatar`
   - Session stores complete user data with avatar
   - Navigation components show avatar immediately

3. **Session Persistence:**
   - Avatar is stored in encrypted JWT session cookie
   - Persists across page reloads and browser sessions
   - No need to re-fetch from backend on every page load

## Components That Display Avatar

- `UserMenu` (header dropdown) - Uses `user.avatar`
- `NavUser` (sidebar user section) - Uses `user.avatar`
- Both components fall back to `getAvatarSrc(user.name)` if avatar is not set

## Testing

1. **Upload Avatar:**
   - Go to Settings → Account Settings
   - Upload a profile picture
   - Avatar should appear immediately in header and sidebar
   - No page reload needed

2. **Login Persistence:**
   - Upload avatar
   - Log out
   - Log back in
   - Avatar should still be visible in header and sidebar

3. **Page Reload:**
   - Upload avatar
   - Reload the page
   - Avatar should persist without re-upload

## Files Modified

### Backend

- `backend/types/auth.go` - Added Preferences field to UserResponse
- `backend/services/auth_service.go` - Updated Login and Register to include preferences

### Frontend

- `frontend/src/app/_actions/auth.ts` - Updated login and registration to extract avatar
- `frontend/src/app/_actions/settings.ts` - Updated profile actions to handle avatar properly

## Build Status

✅ Backend compiles successfully
✅ Frontend builds successfully (52 routes)
✅ No TypeScript errors
✅ No Go compilation errors

## Next Steps

- Test avatar upload and persistence
- Verify avatar shows after login
- Confirm avatar persists across sessions
