# Organization Logo Upload - Usage Examples

## Quick Start

The easiest way to add logo upload to any page is using the `OrganizationLogoSection` component:

```tsx
import { OrganizationLogoSection } from "@/components/organization/organization-logo-section";
import { useOrganizationContext } from "@/hooks/use-organization";

function OrganizationSettingsPage() {
  const { currentOrganization } = useOrganizationContext();

  if (!currentOrganization) return null;

  return (
    <div className="space-y-6">
      <OrganizationLogoSection
        organizationId={currentOrganization.id}
        organizationName={currentOrganization.name}
        currentLogoUrl={currentOrganization.logoUrl}
        onLogoUpdated={(url) => {
          console.log("Logo updated:", url);
          // Optionally refresh organization data
        }}
      />
    </div>
  );
}
```

## Individual Components

### 1. Upload Component Only

For custom save logic:

```tsx
import { useState } from "react";
import { OrganizationLogoUpload } from "@/components/ui/organization-logo-upload";
import { Button } from "@/components/ui/button";

function CustomLogoUpload() {
  const [logoUrl, setLogoUrl] = useState("");

  const handleSave = async () => {
    // Your custom save logic
    await fetch("/api/organizations/123", {
      method: "PUT",
      body: JSON.stringify({ logoUrl }),
    });
  };

  return (
    <div>
      <OrganizationLogoUpload
        currentLogoUrl={logoUrl}
        organizationName="My Organization"
        onLogoChange={setLogoUrl}
        size="md"
      />
      <Button onClick={handleSave}>Save</Button>
    </div>
  );
}
```

### 2. Display Component Only

For showing logos without upload:

```tsx
import { OrganizationAvatar } from "@/components/ui/organization-avatar";

function OrganizationCard({ organization }) {
  return (
    <div className="flex items-center gap-3">
      <OrganizationAvatar
        name={organization.name}
        logoUrl={organization.logoUrl}
        size="lg"
      />
      <div>
        <h3>{organization.name}</h3>
        <p>{organization.description}</p>
      </div>
    </div>
  );
}
```

## Size Variants

### Upload Component Sizes

```tsx
// Small - 64px (h-16 w-16)
<OrganizationLogoUpload size="sm" {...props} />

// Medium - 96px (h-24 w-24) - Default
<OrganizationLogoUpload size="md" {...props} />

// Large - 128px (h-32 w-32)
<OrganizationLogoUpload size="lg" {...props} />
```

### Display Component Sizes

```tsx
// Extra Small - 24px
<OrganizationAvatar size="xs" {...props} />

// Small - 32px
<OrganizationAvatar size="sm" {...props} />

// Medium - 40px - Default
<OrganizationAvatar size="md" {...props} />

// Large - 48px
<OrganizationAvatar size="lg" {...props} />

// Extra Large - 64px
<OrganizationAvatar size="xl" {...props} />
```

## Advanced Usage

### With Form Integration

```tsx
import { useForm } from "react-hook-form";
import { OrganizationLogoUpload } from "@/components/ui/organization-logo-upload";

function OrganizationForm() {
  const { register, handleSubmit, watch, setValue } = useForm({
    defaultValues: {
      name: "",
      description: "",
      logoUrl: "",
    },
  });

  const organizationName = watch("name") || "New Organization";
  const logoUrl = watch("logoUrl");

  const onSubmit = async (data) => {
    // Submit form with logo URL
    console.log(data);
  };

  return (
    <form onSubmit={handleSubmit(onSubmit)}>
      <input {...register("name")} placeholder="Organization Name" />

      <OrganizationLogoUpload
        currentLogoUrl={logoUrl}
        organizationName={organizationName}
        onLogoChange={(url) => setValue("logoUrl", url)}
      />

      <button type="submit">Create Organization</button>
    </form>
  );
}
```

### With Custom Styling

```tsx
<OrganizationAvatar
  name="My Org"
  logoUrl="https://..."
  size="md"
  className="border-2 border-primary"
  fallbackClassName="bg-gradient-to-br from-blue-500 to-purple-600"
/>
```

### In a List/Grid

```tsx
function OrganizationList({ organizations }) {
  return (
    <div className="grid grid-cols-3 gap-4">
      {organizations.map((org) => (
        <div key={org.id} className="p-4 border rounded-lg">
          <OrganizationAvatar
            name={org.name}
            logoUrl={org.logoUrl}
            size="lg"
            className="mx-auto mb-3"
          />
          <h3 className="text-center font-medium">{org.name}</h3>
        </div>
      ))}
    </div>
  );
}
```

### With Loading State

```tsx
function OrganizationProfile() {
  const { data: org, isLoading } = useOrganization();

  if (isLoading) {
    return <div className="h-12 w-12 bg-muted animate-pulse rounded-lg" />;
  }

  return <OrganizationAvatar name={org.name} logoUrl={org.logoUrl} size="lg" />;
}
```

## Common Patterns

### Settings Page

```tsx
function OrganizationSettings() {
  const { currentOrganization } = useOrganizationContext();

  return (
    <div className="max-w-2xl space-y-8">
      {/* Logo Section */}
      <section className="border-b pb-6">
        <OrganizationLogoSection
          organizationId={currentOrganization.id}
          organizationName={currentOrganization.name}
          currentLogoUrl={currentOrganization.logoUrl}
        />
      </section>

      {/* Other settings sections */}
      <section>{/* ... */}</section>
    </div>
  );
}
```

### Profile Header

```tsx
function OrganizationHeader() {
  const { currentOrganization } = useOrganizationContext();

  return (
    <div className="flex items-center gap-4 p-6 bg-card rounded-lg">
      <OrganizationAvatar
        name={currentOrganization.name}
        logoUrl={currentOrganization.logoUrl}
        size="xl"
      />
      <div>
        <h1 className="text-2xl font-bold">{currentOrganization.name}</h1>
        <p className="text-muted-foreground">
          {currentOrganization.description}
        </p>
      </div>
    </div>
  );
}
```

### Dropdown/Select

```tsx
function OrganizationSelector({ organizations, value, onChange }) {
  return (
    <Select value={value} onValueChange={onChange}>
      <SelectTrigger>
        <div className="flex items-center gap-2">
          <OrganizationAvatar
            name={organizations.find((o) => o.id === value)?.name}
            logoUrl={organizations.find((o) => o.id === value)?.logoUrl}
            size="xs"
          />
          <span>{organizations.find((o) => o.id === value)?.name}</span>
        </div>
      </SelectTrigger>
      <SelectContent>
        {organizations.map((org) => (
          <SelectItem key={org.id} value={org.id}>
            <div className="flex items-center gap-2">
              <OrganizationAvatar
                name={org.name}
                logoUrl={org.logoUrl}
                size="xs"
              />
              <span>{org.name}</span>
            </div>
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  );
}
```

## Error Handling

```tsx
function SafeOrganizationAvatar({ organization }) {
  if (!organization) {
    return <div className="h-10 w-10 bg-muted rounded-lg" />;
  }

  return (
    <OrganizationAvatar
      name={organization.name || "Unknown"}
      logoUrl={organization.logoUrl}
      size="md"
    />
  );
}
```

## Accessibility

The components are built with accessibility in mind:

- Proper alt text for images
- Keyboard navigation support
- Screen reader friendly
- Focus indicators
- ARIA labels where appropriate

## Performance Tips

1. Use appropriate sizes - don't load xl images for xs displays
2. The avatar component automatically optimizes images via ImageKit
3. Images are lazy-loaded by default
4. Consider using skeleton loaders for better perceived performance

```tsx
{
  isLoading ? (
    <div className="h-10 w-10 bg-muted animate-pulse rounded-lg" />
  ) : (
    <OrganizationAvatar {...props} />
  );
}
```
