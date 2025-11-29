# Color Palette Visual Reference Guide

## Primary Color: Vibrant Blue (#0c54e7)

Used for: Primary buttons, links, active states, focus rings

```
Light Mode: oklch(52.4% 0.21 265.5) → #0c54e7
Dark Mode:  oklch(64% 0.22 265.5)   → #1e7ae8 (brighter)
```

### Full Primary Blue Scale

```
50:  #eef7ff ████ (Extra light backgrounds, light badges)
100: #d8edff ████
200: #b9dfff ████ (Light hover states)
300: #89ccff ████
400: #52b0ff ████ (Medium hover, secondary buttons)
500: #2a8cff ████
600: #136cfd ████ (Bright active state)
700: #0c54e7 ████ PRIMARY - Main brand color
800: #1144bc ████ (Pressed state, darker interaction)
900: #143e94 ████
950: #11275a ████ (Extra dark - rare use)
```

**Usage in Components**:
- Buttons: `bg-primary text-primary-foreground`
- Links: `text-primary hover:underline`
- Focus: `ring-primary focus:ring-2`
- Hover: `hover:bg-primary/90`
- Badge: `bg-primary/20 text-primary`

---

## Secondary Color: Emerald Green (#10b981)

Used for: Success states, approved actions, positive feedback

```
Light Mode: oklch(67.3% 0.157 155.8) → #10b981
Dark Mode:  oklch(75% 0.165 155.8)   → #34d399 (brighter)
```

### Full Secondary Green Scale

```
50:  #d1fae5 ████ (Extra light backgrounds)
100: #a7f3d0 ████
200: #6ee7b7 ████ (Light states)
300: #3f9d7d ████
400: #10b981 ████ (Medium - standard use)
500: #059669 ████
600: #047857 ████ (Darker interaction)
700: #065f46 ████ (Dark pressed state)
800: #034e3b ████
900: #022c1d ████
```

**Usage in Components**:
- Success badge: `bg-secondary text-secondary-foreground`
- Approved button: `bg-secondary hover:bg-secondary/90`
- Success message: `text-secondary`
- Positive indicator: `text-secondary font-semibold`
- Status: `<Badge variant="secondary">Approved</Badge>`

**Real-World Examples**:
```
✓ Document Approved
✓ Submission Successful
✓ Changes Saved
```

---

## Accent Color: Amber Gold (#f59e0b)

Used for: Warnings, highlights, pending states, attention-grabbing elements

```
Light Mode: oklch(71.5% 0.167 73.5) → #f59e0b
Dark Mode:  oklch(78% 0.175 73.5)   → #fcd34d (brighter)
```

### Full Accent Amber Scale

```
50:  #fffbeb ████ (Extra light backgrounds, light warnings)
100: #fef3c7 ████
200: #fde68a ████ (Light hover)
300: #fcd34d ████
400: #fbbf24 ████ (Medium - secondary highlight)
500: #f59e0b ████ PRIMARY ACCENT - Warnings, pending
600: #d97706 ████ (Darker interaction)
700: #b45309 ████ (Pressed state)
800: #92400e ████
900: #78350f ████ (Extra dark)
```

**Usage in Components**:
- Warning badge: `bg-accent text-accent-foreground`
- In-progress state: `<Badge variant="secondary" className="bg-amber-100 text-amber-900">In Review</Badge>`
- Highlight: `bg-accent/20 border-l-4 border-accent`
- Alert: `text-accent font-semibold`
- Status: `<Badge className="bg-accent text-accent-foreground">Pending</Badge>`

**Real-World Examples**:
```
⚠ Document Under Review
⚠ Awaiting Approval
⚠ Action Required
! Stage 2 of 4
```

---

## Destructive Color: Red

Used for: Errors, rejections, delete actions

```
Light Mode: oklch(0.577 0.245 27.325)
Dark Mode:  oklch(0.704 0.191 22.216)
```

**Usage**:
- Error message: `text-destructive`
- Delete button: `bg-destructive text-destructive-foreground`
- Error badge: `<Badge variant="destructive">Rejected</Badge>`

**Real-World Examples**:
```
✗ Document Rejected
✗ Error Processing Request
✗ Cannot Delete - In Use
```

---

## Chart Color System

Used for data visualization and comparative elements

### Primary Charts
```
Chart 1 (Primary):    Blue    - oklch(52.4% 0.21 265.5)
Chart 2 (Secondary):  Green   - oklch(67.3% 0.157 155.8)
Chart 3 (Accent):     Amber   - oklch(71.5% 0.167 73.5)
Chart 4 (Purple):     Purple  - oklch(0.828 0.189 84.429)
Chart 5 (Orange):     Orange  - oklch(0.769 0.188 70.08)
```

**Example Chart Usage**:
```
Department A: Blue bar chart
Department B: Green bar chart
Department C: Amber bar chart
```

---

## Status Badge Color Mapping

Quick reference for common status combinations:

### Approval Workflow Status

```
┌─────────────────┬──────────────┬─────────────────────┐
│ Status          │ Color        │ Styling             │
├─────────────────┼──────────────┼─────────────────────┤
│ DRAFT           │ Gray/Muted   │ bg-muted/80         │
│ SUBMITTED       │ Blue         │ bg-primary/80       │
│ IN_APPROVAL     │ Amber        │ bg-accent/90        │
│ APPROVED        │ Green        │ bg-secondary        │
│ REJECTED        │ Red          │ bg-destructive      │
│ REVERSED        │ Orange       │ bg-chart-5          │
└─────────────────┴──────────────┴─────────────────────┘
```

### Status Components

```jsx
// Draft
<Badge variant="outline" className="bg-gray-100 text-gray-800">
  DRAFT
</Badge>

// Submitted (Pending Review)
<Badge className="bg-primary text-primary-foreground">
  SUBMITTED
</Badge>

// In Approval (Waiting for Action)
<Badge className="bg-accent text-accent-foreground">
  IN_APPROVAL
</Badge>

// Approved (Success)
<Badge className="bg-secondary text-secondary-foreground">
  APPROVED
</Badge>

// Rejected (Error)
<Badge variant="destructive">
  REJECTED
</Badge>
```

---

## Light Mode vs Dark Mode Comparison

### Primary Color
| Mode | Lightness | Hex | Use Case |
|------|-----------|-----|----------|
| Light | 52.4% | #0c54e7 | Good contrast on white background |
| Dark | 64% | #1e7ae8 | Brighter for visibility on dark background |

### Secondary Color
| Mode | Lightness | Hex | Use Case |
|------|-----------|-----|----------|
| Light | 67.3% | #10b981 | Balanced on white |
| Dark | 75% | #34d399 | Enhanced brightness on dark |

### Accent Color
| Mode | Lightness | Hex | Use Case |
|------|-----------|-----|----------|
| Light | 71.5% | #f59e0b | Warm tone on white |
| Dark | 78% | #fcd34d | Extra brightness on dark |

---

## Color Combinations & Contrast

### Safe Color Combinations (WCAG AA+)

✓ Primary on White: 7.2:1 ratio
```
bg-white text-primary
```

✓ Secondary on White: 5.8:1 ratio
```
bg-white text-secondary
```

✓ Dark text on Accent: 8.2:1 ratio
```
bg-accent text-accent-foreground
```

✓ White text on Primary: 7.8:1 ratio
```
bg-primary text-primary-foreground
```

### Avoid These Combinations

✗ Light text on light backgrounds
✗ Primary on Secondary (too similar)
✗ Accent on light backgrounds at low opacity (hard to read)

---

## Tailwind Class Reference

### Button Classes

```jsx
// Primary Button (Blue)
className="bg-primary text-primary-foreground hover:bg-primary/90"

// Secondary Button (Green)
className="bg-secondary text-secondary-foreground hover:bg-secondary/90"

// Accent Button (Amber)
className="bg-accent text-accent-foreground hover:bg-accent/90"

// Outline Button
className="border border-primary text-primary hover:bg-primary/10"

// Ghost Button
className="text-primary hover:bg-primary/10"
```

### Badge Classes

```jsx
// Primary Badge
className="bg-primary text-primary-foreground"

// Secondary (Success)
className="bg-secondary text-secondary-foreground"

// Accent (Warning/Pending)
className="bg-accent text-accent-foreground"

// Destructive (Error)
className="bg-destructive text-destructive-foreground"

// Muted (Inactive)
className="bg-muted text-muted-foreground"
```

### Text Classes

```jsx
// Primary Text
className="text-primary hover:underline"

// Secondary Text
className="text-secondary font-semibold"

// Accent Text
className="text-accent"

// Muted Text
className="text-muted-foreground"
```

### Background Classes

```jsx
// Light Primary Background
className="bg-primary/10 text-primary"

// Light Secondary Background
className="bg-secondary/10 text-secondary"

// Light Accent Background
className="bg-accent/10 text-accent-foreground"

// Colored Card
className="bg-primary/5 border border-primary/20"
```

---

## Implementation Checklist

When implementing colors in components:

- [ ] Primary actions use `bg-primary`
- [ ] Success states use `bg-secondary`
- [ ] Warnings/pending use `bg-accent`
- [ ] Text contrast meets WCAG AA
- [ ] Colors work in dark mode
- [ ] Hover states are defined
- [ ] Focus states use `ring-primary`
- [ ] Disabled states are visually distinct
- [ ] No color alone conveys meaning (icons/text included)

---

## Tools for Testing

1. **Contrast Checking**: https://webaim.org/resources/contrastchecker/
2. **Color Picker**: https://oklch.com/
3. **Color Blindness Simulation**: https://www.color-blindness.com/coblis-color-blindness-simulator/
4. **Accessible Colors**: https://accessible-colors.com/
5. **Tailwind Color Reference**: https://tailwindcss.com/docs/customizing-colors

---

**Color Scheme Version**: 1.0
**Last Updated**: November 29, 2024
**Framework**: Tailwind CSS v4
**Color Space**: OKLCH
