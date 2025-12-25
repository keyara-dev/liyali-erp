# Phase 2: Frontend Testing Guide for Signup Component

## Test Environment Setup

### Prerequisites
- Frontend running on `http://localhost:3001`
- Backend running on `http://localhost:8080`
- Database initialized and accessible
- Browser DevTools open (F12)

### Frontend Setup Checklist
- [ ] `.env` file has `BASE_URL=http://localhost:8080`
- [ ] Backend API is accessible from frontend
- [ ] No CORS errors in browser console
- [ ] Session/cookies enabled in browser

---

## Component Structure Review

**File**: `frontend/src/app/(auth)/_components/signup.tsx`

### Form Fields Implemented
- ✅ Email input with validation
- ✅ Full Name input
- ✅ Password input with show/hide toggle
- ✅ Confirm Password input with show/hide toggle
- ✅ Password strength requirements (displayed below password field)
- ✅ Error message display with AlertCircle icon
- ✅ Loading state on submit button
- ✅ "Already have an account? Login" link

### Password Validation Rules
- Minimum 8 characters
- At least 1 uppercase letter (A-Z)
- At least 1 lowercase letter (a-z)
- At least 1 digit (0-9)

---

## Test Case 1: Form Renders Correctly ✅

### Objective
Verify that signup form displays all fields and UI elements correctly.

### Steps
1. Navigate to `http://localhost:3001/signup`
2. Wait for page to load
3. Inspect all form elements

### Verification Checklist
- [ ] Page title displays: "Create your **account**" (with bold on "account")
- [ ] Subtitle displays: "Join us and start using Liyali Gateway in minutes."
- [ ] Email input field visible with placeholder "you@example.com"
- [ ] Full Name input field visible with placeholder "John Doe"
- [ ] Password input field visible with placeholder "Enter a strong password"
- [ ] Password requirements text visible: "Must have: 8+ characters, 1 uppercase, 1 lowercase, 1 digit"
- [ ] Eye icon button appears when password is entered
- [ ] Confirm Password input field visible with placeholder "Confirm your password"
- [ ] Eye icon button appears when confirm password is entered
- [ ] Submit button visible with text "Sign Up"
- [ ] Submit button is **not disabled** initially
- [ ] Bottom link displays: "Already have an account? Login" (Login is a link)
- [ ] No error messages visible initially

### Browser Console Check
- [ ] No JavaScript errors
- [ ] No API errors
- [ ] No styling issues (check layout in DevTools)

---

## Test Case 2: Password Visibility Toggle Works ✅

### Objective
Verify that show/hide password toggles work correctly for both fields.

### Steps
1. Navigate to signup page
2. Enter password in password field: `TestPass123`
3. Click eye icon next to password field
4. Verify password visibility changes
5. Click eye icon again
6. Repeat for confirm password field

### Verification Checklist

#### Password Field Toggle
- [ ] When empty, eye icon is **not displayed**
- [ ] When text entered, eye icon **appears**
- [ ] Eye icon shows "eye-off" symbol initially (password hidden)
- [ ] Clicking eye icon changes to "eye" symbol (password visible)
- [ ] Password text becomes visible when toggled
- [ ] Password text becomes hidden when toggled back
- [ ] Toggle works multiple times

#### Confirm Password Field Toggle
- [ ] Same behavior as password field
- [ ] Both fields toggle independently
- [ ] Toggling one doesn't affect the other

---

## Test Case 3: Form Validation - Empty Fields ❌

### Objective
Verify that form rejects submission with empty fields.

### Steps
1. Navigate to signup page
2. Click "Sign Up" button without entering any data
3. Observe error handling

### Verification Checklist
- [ ] Form submission is **prevented** (no API call made)
- [ ] Error message displayed: "Please fill in all required fields"
- [ ] Error message appears in red box with AlertCircle icon
- [ ] Form is not submitted
- [ ] Browser console shows no API errors

### Test Each Empty Field
```
Email empty, others filled:
  Expected: Validation error

Name empty, others filled:
  Expected: Validation error

Password empty, others filled:
  Expected: Validation error

Confirm password empty, others filled:
  Expected: Validation error
```

---

## Test Case 4: Password Validation - Strength Requirements ❌

### Objective
Verify that weak passwords are rejected with detailed error messages.

### Test Case 4.1: Password Too Short
**Steps**
1. Email: `short@example.com`
2. Name: `Short Pass`
3. Password: `Pass1` (only 5 characters)
4. Confirm Password: `Pass1`
5. Click "Sign Up"

**Expected**
- [ ] Form submission prevented
- [ ] Error message: "Password requirements: At least 8 characters"
- [ ] No API call made
- [ ] Form remains populated with entered data (except password fields?)

### Test Case 4.2: No Uppercase Letter
**Steps**
1. Email: `noupper@example.com`
2. Name: `No Upper`
3. Password: `password123` (all lowercase)
4. Confirm Password: `password123`
5. Click "Sign Up"

**Expected**
- [ ] Error message: "Password requirements: At least 1 uppercase letter"
- [ ] Form submission prevented
- [ ] No API call made

### Test Case 4.3: No Lowercase Letter
**Steps**
1. Email: `nolower@example.com`
2. Name: `No Lower`
3. Password: `PASSWORD123` (all uppercase)
4. Confirm Password: `PASSWORD123`
5. Click "Sign Up"

**Expected**
- [ ] Error message: "Password requirements: At least 1 lowercase letter"
- [ ] Form submission prevented

### Test Case 4.4: No Digit
**Steps**
1. Email: `nodigit@example.com`
2. Name: `No Digit`
3. Password: `PasswordOnly` (no numbers)
4. Confirm Password: `PasswordOnly`
5. Click "Sign Up"

**Expected**
- [ ] Error message: "Password requirements: At least 1 digit"
- [ ] Form submission prevented

### Test Case 4.5: Multiple Violations
**Steps**
1. Email: `multiple@example.com`
2. Name: `Multiple`
3. Password: `pass` (short + no uppercase + no digit)
4. Confirm Password: `pass`
5. Click "Sign Up"

**Expected**
- [ ] Error message includes **all violations**: "Password requirements: At least 8 characters, At least 1 uppercase letter, At least 1 digit"
- [ ] Form submission prevented

---

## Test Case 5: Password Mismatch Validation ❌

### Objective
Verify that passwords must match.

### Steps
1. Email: `nomatch@example.com`
2. Name: `No Match`
3. Password: `ValidPass123`
4. Confirm Password: `ValidPass456` (different)
5. Click "Sign Up"

### Verification Checklist
- [ ] Form submission prevented
- [ ] Error message: "Passwords do not match. Please try again."
- [ ] No API call made
- [ ] User remains on form

---

## Test Case 6: Successful Registration - Happy Path ✅

### Objective
Verify complete successful registration flow.

### Steps
1. Navigate to `http://localhost:3001/signup`
2. Enter data:
   - Email: `newuser+test@example.com`
   - Name: `New Test User`
   - Password: `SecurePassword123`
   - Confirm Password: `SecurePassword123`
3. Click "Sign Up"
4. Observe submission process and redirect

### Verification Checklist

#### Form Submission
- [ ] Submit button becomes **disabled** during submission
- [ ] Submit button text changes to "Creating account..."
- [ ] Form fields become **disabled** (grayed out)
- [ ] No user input accepted while loading

#### API Call
- [ ] Browser Network tab shows POST request to `/api/v1/auth/register`
- [ ] Request includes headers:
  - `Content-Type: application/json`
- [ ] Request body contains:
  ```json
  {
    "email": "newuser+test@example.com",
    "name": "New Test User",
    "password": "SecurePassword123",
    "role": "requester"
  }
  ```

#### Response Handling
- [ ] Backend returns 201 Created status
- [ ] Response includes:
  - `success: true`
  - `token: "..."` (valid JWT)
  - `user: {...}`
  - `organization: {...}`
- [ ] No error message displays
- [ ] Session is created (check DevTools → Application → Cookies)

#### Session Creation
- [ ] Cookie `AUTH_SESSION` is created
- [ ] Cookie contains encrypted token
- [ ] Cookie is httpOnly (cannot be accessed via JavaScript)
- [ ] Session includes:
  - `access_token` (from response.token)
  - `role` (from response.user.role = "requester")
  - `user_id` (from response.user.id)
  - `organization_id` (from response.organization.id) ← **IMPORTANT**

#### Redirect
- [ ] Page automatically redirects to `/home` (dashboard)
- [ ] Redirect happens **after** session is created
- [ ] User is logged in on dashboard (can see their name/email)
- [ ] No intermediate screens appear
- [ ] URL changes from `/signup` to `/home`

#### Browser Console
- [ ] No JavaScript errors
- [ ] No console warnings
- [ ] No failed API calls

---

## Test Case 7: Duplicate Email Registration ❌

### Objective
Verify that registering with existing email shows proper error.

### Process
1. Register user with email: `duplicate@example.com` ← Success
2. Try to register again with same email

### Steps for Second Registration
1. Navigate to `http://localhost:3001/signup`
2. Email: `duplicate@example.com` (same as first user)
3. Name: `Different Name`
4. Password: `DifferentPass456`
5. Confirm Password: `DifferentPass456`
6. Click "Sign Up"

### Verification Checklist
- [ ] Form submission happens (button disabled, loading state shown)
- [ ] API call made to backend
- [ ] Backend returns **409 Conflict** status
- [ ] Error message displayed: Shows backend message
- [ ] Submit button becomes **enabled** again (ready to retry)
- [ ] Form fields become **enabled** again
- [ ] User remains on signup page
- [ ] **No redirect** to `/home`
- [ ] Session is **not created**
- [ ] No AUTH_SESSION cookie added

---

## Test Case 8: Backend Validation Errors ❌

### Objective
Verify proper error handling when backend validation fails.

### Test Case 8.1: Invalid Role (Should Not Happen - Frontend Doesn't Allow)
If somehow role is set to invalid value in code:
- [ ] Backend rejects with 400 error
- [ ] Frontend displays error message
- [ ] Retry possible

### Test Case 8.2: Server Error (500)
Simulate backend error:
- [ ] Error message displays appropriately
- [ ] Button becomes enabled for retry
- [ ] Form remains populated
- [ ] Console shows error details

### Test Case 8.3: Network Error
1. Open DevTools → Network
2. Check "Offline" to simulate network failure
3. Try to register
4. Expected:
   - [ ] Request fails
   - [ ] Error message displays
   - [ ] Button becomes enabled
   - [ ] Form remains populated

---

## Test Case 9: Form Data Persistence on Error ❌

### Objective
Verify that form data is preserved when errors occur.

### Steps
1. Enter all form data:
   - Email: `persist@example.com`
   - Name: `Persist Test`
   - Password: `ValidPass123`
   - Confirm Password: `ValidPass123`
2. Trigger an error (e.g., network offline or duplicate email)
3. Check form fields

### Verification Checklist
- [ ] Email field still contains: `persist@example.com`
- [ ] Name field still contains: `Persist Test`
- [ ] Password field **empty or shows placeholder** (security: don't show passwords in retry)
- [ ] Confirm Password field **empty or shows placeholder**
- [ ] Error message displays
- [ ] User can read error and potentially fix it

---

## Test Case 10: Login Link Navigation ✅

### Objective
Verify that "Already have an account? Login" link works.

### Steps
1. Navigate to `http://localhost:3001/signup`
2. Click on "Login" link at bottom
3. Observe navigation

### Verification Checklist
- [ ] Page navigates to `/login`
- [ ] Login form displays
- [ ] URL shows `/login`
- [ ] No errors in console

---

## Test Case 11: Response Storage in Session ✅

### Objective
Verify that organization data from registration response is properly stored.

### Steps
1. Register new user successfully
2. Open DevTools → Console
3. Run JavaScript to check session:
   ```javascript
   // Note: Session is httpOnly, so we can't access it directly
   // But we can verify the organization switcher works
   ```
4. Check that organization switcher displays user's organization

### Verification Checklist (Post-Registration on Dashboard)
- [ ] Organization switcher shows the created organization
- [ ] Organization name matches user's name (from registration)
- [ ] Can switch organizations (if user added to others)
- [ ] Organization context flows through API calls
- [ ] `X-Organization-ID` header includes org from registration

---

## Test Case 12: Password Not Logged/Displayed ✅

### Objective
Verify that passwords are never logged or displayed in console.

### Steps
1. Open DevTools → Console
2. Register with password: `SecurePassword123`
3. Check Network tab for API request
4. Check console logs

### Verification Checklist
- [ ] Console does **not** log password anywhere
- [ ] Network request body **contains** password (encrypted in transit via HTTPS)
- [ ] Password not logged in error messages
- [ ] No sensitive data in DevTools console
- [ ] No password in localStorage or sessionStorage (only token)

---

## Test Case 13: Email Format Validation ✅

### Objective
Verify that invalid email formats are caught.

### Test Cases
```
blank@:          → Browser validation (HTML5 type="email")
nodomain@        → Browser validation
@nodomain.com    → Browser validation
spaces in@email  → Browser validation
```

### Expected Behavior
- [ ] HTML5 form validation prevents submission
- [ ] Browser shows error message
- [ ] Form doesn't submit to backend
- [ ] User can correct email and retry

---

## Test Case 14: Accessibility - Tab Navigation ✅

### Objective
Verify form is keyboard navigable.

### Steps
1. Navigate to signup page
2. Press Tab key multiple times
3. Observe focus order

### Expected Focus Order
1. Email input
2. Name input
3. Password input
4. Show/hide password button
5. Confirm Password input
6. Show/hide confirm password button
7. Sign Up button
8. Login link

### Verification Checklist
- [ ] Tab key moves focus through fields in logical order
- [ ] All interactive elements are focusable
- [ ] Focus indicator (outline) visible on all elements
- [ ] Can submit form using keyboard (Tab to button, Enter to submit)
- [ ] No focus traps (can always move forward with Tab)

---

## Test Case 15: Loading State Visual Feedback ✅

### Objective
Verify that loading state provides clear visual feedback.

### Steps
1. Start registration
2. Observe button state during submission
3. Add network throttling to see longer load time:
   - DevTools → Network → Slow 3G
4. Register and observe feedback

### Verification Checklist
- [ ] Submit button text changes to "Creating account..."
- [ ] Submit button becomes visually disabled (grayed out/opacity change)
- [ ] Submit button shows loading indicator (optional: spinner)
- [ ] Form fields become disabled
- [ ] User cannot submit twice simultaneously
- [ ] After response, button text reverts to "Sign Up"
- [ ] Form fields become enabled

---

## Test Case 16: Responsive Design ✅

### Objective
Verify form works on different screen sizes.

### Test on Different Viewports
```
Desktop (1920x1080)
  → [ ] Form displays properly
  → [ ] No horizontal scroll
  → [ ] All fields visible
  → [ ] Button clickable

Tablet (768x1024)
  → [ ] Form displays properly
  → [ ] Input fields adequately sized
  → [ ] Button easily clickable
  → [ ] No layout issues

Mobile (375x667)
  → [ ] Form stacks vertically
  → [ ] Inputs take full width (with padding)
  → [ ] Password toggle buttons accessible
  → [ ] Button spans full width
  → [ ] No horizontal scroll
  → [ ] Text readable (font size >= 16px)
```

### DevTools Testing
1. Press F12 → Toggle device toolbar (Ctrl+Shift+M)
2. Test on different device presets
3. Verify responsiveness

---

## Test Case 17: Success Message Clarity ✅

### Objective
Verify that successful registration is clear.

### Steps
1. Register successfully
2. Observe any on-screen feedback before redirect

### Verification Checklist
- [ ] No error message displayed
- [ ] Page immediately redirects (within 1-2 seconds)
- [ ] Loading state shows "Creating account..."
- [ ] Dashboard loads with user's data

---

## Test Case 18: Integration - Session Persists ✅

### Objective
Verify that session created during signup persists across pages.

### Steps
1. Register and get redirected to `/home`
2. Refresh page (F5)
3. Check that user is still logged in

### Verification Checklist
- [ ] Page doesn't redirect to login after refresh
- [ ] User data displays correctly
- [ ] Session cookie persists (check DevTools → Application)
- [ ] Can navigate to other protected pages
- [ ] Logout works and clears session

---

## Test Case 19: Security - No Plaintext Passwords in Requests ✅

### Objective
Verify that passwords are transmitted securely.

### Steps
1. Open DevTools → Network
2. Filter for XHR/Fetch requests
3. Register and check request

### Verification Checklist
- [ ] Request uses POST method
- [ ] Connection is HTTPS (in production) or HTTP:// for localhost
- [ ] Request body shows password field (plaintext in request is OK over HTTPS)
- [ ] Response doesn't include password
- [ ] No password in URL or headers

---

## Test Case 20: Error Recovery - Can Register After Failed Attempt ✅

### Objective
Verify that user can correct errors and retry.

### Steps
1. Try to register with weak password
2. See error message
3. Fix password
4. Try again
5. Should succeed

### Verification Checklist
- [ ] First attempt fails with appropriate error
- [ ] Error message shows what's wrong
- [ ] User can modify form (fields are enabled)
- [ ] Second attempt succeeds
- [ ] User redirected to `/home`

---

## Manual Testing Checklist

### Before Running Tests
- [ ] Backend is running on `http://localhost:8080`
- [ ] Frontend is running on `http://localhost:3001`
- [ ] Database is initialized
- [ ] Browser cache cleared (or use incognito mode)
- [ ] DevTools opened on separate monitor (or split screen)

### Quick Test Sequence
```
1. [ ] Navigate to /signup
2. [ ] Form displays correctly
3. [ ] Try weak password → error shown
4. [ ] Try mismatch passwords → error shown
5. [ ] Enter valid data
6. [ ] Submit
7. [ ] See loading state
8. [ ] Redirected to /home
9. [ ] Logged in successfully
10. [ ] Logout and verify session cleared
```

### Test Result Log
```
Date: 2025-12-25
Tester: [Name]
Environment: Local Dev

Test Results:
- Form Rendering: PASS / FAIL
- Password Validation: PASS / FAIL
- Form Submission: PASS / FAIL
- Session Creation: PASS / FAIL
- Redirect: PASS / FAIL
- Error Handling: PASS / FAIL
- Security: PASS / FAIL

Issues Found:
1. [Issue Description]
   Severity: Low / Medium / High
   Resolution:

Overall Status: PASS / FAIL
```

---

## Automated Testing (Jest/React Testing Library)

### Example Test File Location
`frontend/src/app/(auth)/_components/__tests__/signup.test.tsx`

### Example Test Cases (To Be Written)
```typescript
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import Signup from '../signup';

describe('Signup Component', () => {
  it('renders signup form with all fields', () => {
    render(<Signup />);
    expect(screen.getByLabelText(/email/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/full name/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/^password$/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/confirm password/i)).toBeInTheDocument();
  });

  it('disables submit button when passwords do not match', async () => {
    render(<Signup />);
    await userEvent.type(screen.getByLabelText(/^password$/i), 'TestPass123');
    await userEvent.type(screen.getByLabelText(/confirm password/i), 'Different456');
    expect(screen.getByRole('button', { name: /sign up/i })).toBeDisabled();
  });

  // More tests...
});
```

---

**Status**: Ready for manual testing
**Frontend Tests Created**: 20 test cases
**Estimated Testing Time**: 1-2 hours (manual)
**Date**: 2025-12-25

