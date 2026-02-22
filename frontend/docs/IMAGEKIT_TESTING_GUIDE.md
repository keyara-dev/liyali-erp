# ImageKit Integration Testing Guide

## Pre-Testing Setup

### 1. Get ImageKit Credentials

1. Sign up at [ImageKit.io](https://imagekit.io/)
2. Navigate to Developer Options → API Keys
3. Copy your:
   - Public Key
   - Private Key
   - URL Endpoint

### 2. Configure Environment

Create/update `frontend/.env.local`:

```env
NEXT_PUBLIC_IMAGEKIT_PUBLIC_KEY=public_xxx
NEXT_PUBLIC_IMAGEKIT_URL_ENDPOINT=https://ik.imagekit.io/your_id
IMAGEKIT_PRIVATE_KEY=private_xxx
NEXT_PUBLIC_IMAGEKIT_AUTH_ENDPOINT=/api/imagekit-auth
```

### 3. Restart Development Server

```bash
cd frontend
npm run dev
```

## Test Scenarios

### Test 1: Authentication Endpoint

**Objective**: Verify the auth endpoint works

```bash
curl http://localhost:3000/api/imagekit-auth
```

**Expected Response**:

```json
{
  "token": "...",
  "expire": 1234567890,
  "signature": "..."
}
```

**Troubleshooting**:

- If 500 error: Check environment variables are set
- If token/signature missing: Verify IMAGEKIT_PRIVATE_KEY is correct

### Test 2: Create Organization with Logo

**Steps**:

1. Navigate to `/welcome` or click "Create workspace"
2. Fill in organization name: "Test Org"
3. Click "Upload Logo" or drag an image
4. Select a valid image (JPG/PNG, < 10MB)
5. Wait for upload progress
6. Click "Create Workspace"

**Expected Behavior**:

- ✅ Upload progress shows 0-100%
- ✅ Preview appears after upload
- ✅ Success toast: "Logo uploaded successfully"
- ✅ Organization created with logo
- ✅ Logo appears in workspace switcher

**Troubleshooting**:

- Upload fails: Check browser console for errors
- No progress: Verify ImageKit credentials
- Image not saving: Check backend receives logoUrl

### Test 3: File Validation

**Test Invalid File Type**:

1. Try uploading a .txt or .pdf file
2. **Expected**: Error toast "Invalid file type..."

**Test Large File**:

1. Try uploading image > 10MB
2. **Expected**: Error toast "File size exceeds 10MB..."

**Test Valid File**:

1. Upload JPG, PNG, GIF, or WebP < 10MB
2. **Expected**: Upload succeeds

### Test 4: Drag and Drop

**Steps**:

1. Open create organization form
2. Drag an image file over the upload area
3. Drop the file

**Expected Behavior**:

- ✅ Drop zone highlights on drag over
- ✅ Upload starts on drop
- ✅ Progress indicator appears

### Test 5: Remove Logo

**Steps**:

1. Upload a logo
2. Click "Remove" button
3. Verify logo is removed

**Expected Behavior**:

- ✅ Logo preview disappears
- ✅ Success toast: "Logo removed"
- ✅ Falls back to initials

### Test 6: Logo Display in Workspace Switcher

**Steps**:

1. Create organization with logo
2. Open workspace switcher (sidebar)
3. Check logo appears

**Expected Behavior**:

- ✅ Logo displays in current workspace button
- ✅ Logo displays in dropdown list
- ✅ Falls back to initials if no logo
- ✅ Images are optimized (check Network tab)

### Test 7: Image Optimization

**Steps**:

1. Upload a large image (e.g., 2000x2000px)
2. Open browser DevTools → Network tab
3. Check the image request

**Expected Behavior**:

- ✅ Image URL contains transformation parameters
- ✅ Image is resized to appropriate dimensions
- ✅ Format is optimized (WebP if supported)
- ✅ File size is reduced

**Example URL**:

```
https://ik.imagekit.io/your_id/tr:w-64,h-64,q-80,f-auto,c-maintain_ratio/organizations/image.jpg
```

### Test 8: Multiple Organizations

**Steps**:

1. Create 3 organizations with different logos
2. Switch between them
3. Verify correct logo displays

**Expected Behavior**:

- ✅ Each organization shows its own logo
- ✅ Switching updates logo immediately
- ✅ No logo mixing/caching issues

### Test 9: Offline Behavior

**Steps**:

1. Open DevTools → Network tab
2. Set to "Offline"
3. Try uploading a logo

**Expected Behavior**:

- ✅ Error toast: "Network error during upload"
- ✅ Upload progress stops
- ✅ Form remains usable

### Test 10: Concurrent Uploads

**Steps**:

1. Open two browser tabs
2. Start upload in both simultaneously
3. Verify both complete

**Expected Behavior**:

- ✅ Both uploads succeed
- ✅ No conflicts or errors
- ✅ Each gets unique authentication

## Browser Testing

Test in multiple browsers:

- [ ] Chrome/Edge (Chromium)
- [ ] Firefox
- [ ] Safari
- [ ] Mobile Safari (iOS)
- [ ] Mobile Chrome (Android)

## Performance Testing

### Upload Speed

**Test with different file sizes**:

- 100KB image: Should complete in < 2 seconds
- 1MB image: Should complete in < 5 seconds
- 5MB image: Should complete in < 15 seconds

**Note**: Times vary based on internet speed

### Image Loading

**Check optimized loading**:

1. Clear browser cache
2. Reload page with organization logos
3. Check Network tab

**Expected**:

- Images load in < 1 second
- Proper caching headers
- Progressive loading

## Security Testing

### Test 1: Private Key Not Exposed

**Steps**:

1. Open browser DevTools → Sources
2. Search for "IMAGEKIT_PRIVATE_KEY"

**Expected**: Should NOT find private key in client code

### Test 2: Token Expiration

**Steps**:

1. Get auth token from `/api/imagekit-auth`
2. Wait 1 hour
3. Try using expired token

**Expected**: Upload should fail with authentication error

### Test 3: Invalid Signature

**Steps**:

1. Intercept upload request
2. Modify signature parameter
3. Send request

**Expected**: ImageKit rejects with 401 error

## Common Issues and Solutions

### Issue: "Failed to get authentication parameters"

**Causes**:

- Missing environment variables
- Incorrect private key
- Auth endpoint not accessible

**Solution**:

```bash
# Check environment variables
echo $IMAGEKIT_PRIVATE_KEY
echo $NEXT_PUBLIC_IMAGEKIT_PUBLIC_KEY

# Restart dev server
npm run dev
```

### Issue: Upload succeeds but image doesn't display

**Causes**:

- URL not saved to database
- CORS issues
- Incorrect URL endpoint

**Solution**:

1. Check browser console for errors
2. Verify logoUrl in database
3. Check ImageKit dashboard for uploaded file

### Issue: Slow uploads

**Causes**:

- Large file size
- Slow internet connection
- ImageKit service issues

**Solution**:

1. Compress images before upload
2. Check internet speed
3. Check ImageKit status page

### Issue: Images not optimized

**Causes**:

- Transformation parameters not applied
- Using wrong URL

**Solution**:

1. Verify using `OrganizationAvatar` component
2. Check URL contains `tr:` parameters
3. Ensure URL is from ImageKit endpoint

## Monitoring

### ImageKit Dashboard

Monitor in ImageKit dashboard:

- Upload count
- Bandwidth usage
- Storage usage
- Error rate

### Application Logs

Check for errors:

```bash
# Frontend logs
npm run dev

# Check browser console
# Look for ImageKit-related errors
```

## Production Checklist

Before deploying to production:

- [ ] Environment variables set in production
- [ ] Private key is secure (not in git)
- [ ] Test uploads in production environment
- [ ] Verify CDN delivery works
- [ ] Check CORS configuration
- [ ] Monitor ImageKit usage
- [ ] Set up alerts for quota limits
- [ ] Test image optimization
- [ ] Verify mobile experience
- [ ] Check accessibility
- [ ] Load test with multiple concurrent uploads

## Support

If issues persist:

1. Check [ImageKit Documentation](https://docs.imagekit.io/)
2. Review browser console errors
3. Check ImageKit dashboard for errors
4. Contact ImageKit support
5. Review implementation in `frontend/src/lib/imagekit.ts`
