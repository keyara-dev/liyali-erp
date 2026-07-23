"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import { format } from "date-fns";
import { toast } from "sonner";
import {
  DndContext,
  DragOverlay,
  type DragEndEvent,
  type DragStartEvent,
  PointerSensor,
  closestCenter,
  useSensor,
  useSensors,
  useDraggable,
  useDroppable,
} from "@dnd-kit/core";
import { CSS } from "@dnd-kit/utilities";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { SelectField } from "@/components/ui/select-field";
import {
  AlertTriangle,
  Building2,
  Calendar,
  CheckCircle2,
  Clock3,
  CircleDot,
  Filter,
  GripVertical,
  Plus,
  RefreshCw,
  Ticket,
  User,
} from "lucide-react";
import { useAdminUsers } from "@/hooks/use-admin-users";
import {
  useCreateSupportTicket,
  useSupportTicketStats,
  useSupportTickets,
  useUpdateSupportTicket,
} from "@/hooks/use-support";
import type {
  CreateSupportTicketRequest,
  SupportTicket,
  SupportTicketFilters,
  UpdateSupportTicketRequest,
} from "@/app/_actions/support";

const statusOptions = [
  { value: "all", label: "All statuses" },
  { value: "open", label: "Open" },
  { value: "in_progress", label: "In Progress" },
  { value: "waiting_on_customer", label: "Waiting on Customer" },
  { value: "resolved", label: "Resolved" },
  { value: "closed", label: "Closed" },
];

const priorityOptions = [
  { value: "all", label: "All priorities" },
  { value: "low", label: "Low" },
  { value: "medium", label: "Medium" },
  { value: "high", label: "High" },
  { value: "urgent", label: "Urgent" },
];

const sourceOptions = [
  { value: "all", label: "All sources" },
  { value: "manual", label: "Manual" },
  { value: "user_app", label: "User App" },
  { value: "email", label: "Email" },
];

const categoryOptions = [
  { value: "general", label: "General" },
  { value: "access", label: "Access" },
  { value: "billing", label: "Billing" },
  { value: "workflow", label: "Workflow" },
  { value: "bug", label: "Bug" },
  { value: "feature_request", label: "Feature Request" },
];

const defaultCreateForm: CreateSupportTicketRequest = {
  subject: "",
  description: "",
  category: "general",
  priority: "medium",
  organization_id: "",
  user_id: "",
  assigned_to_admin_id: "",
  external_reference: "",
  internal_notes: "",
};

type CreateFormState = CreateSupportTicketRequest;

type EditFormState = UpdateSupportTicketRequest & {
  id: string;
};

function getStatusBadge(status: SupportTicket["status"]) {
  switch (status) {
    case "open":
      return <Badge variant="secondary">Open</Badge>;
    case "in_progress":
      return <Badge variant="default">In Progress</Badge>;
    case "waiting_on_customer":
      return <Badge variant="outline">Waiting</Badge>;
    case "resolved":
      return <Badge variant="outline" className="border-green-500 text-green-700">Resolved</Badge>;
    case "closed":
      return <Badge variant="destructive">Closed</Badge>;
    default:
      return <Badge variant="outline">{status}</Badge>;
  }
}

function getPriorityBadge(priority: SupportTicket["priority"]) {
  switch (priority) {
    case "urgent":
      return <Badge variant="destructive">Urgent</Badge>;
    case "high":
      return <Badge variant="destructive" className="bg-orange-100 text-orange-800 hover:bg-orange-100">High</Badge>;
    case "medium":
      return <Badge variant="secondary">Medium</Badge>;
    case "low":
      return <Badge variant="outline">Low</Badge>;
    default:
      return <Badge variant="outline">{priority}</Badge>;
  }
}

function getSourceBadge(source: SupportTicket["source"]) {
  switch (source) {
    case "manual":
      return <Badge variant="outline">Manual</Badge>;
    case "user_app":
      return <Badge variant="secondary">User App</Badge>;
    case "email":
      return <Badge variant="default">Email</Badge>;
    default:
      return <Badge variant="outline">{source}</Badge>;
  }
}

const BOARD_STATUSES: Array<{
  status: SupportTicket["status"];
  title: string;
  description: string;
  accent: string;
}> = [
  {
    status: "open",
    title: "Open",
    description: "New or untriaged tickets",
    accent: "border-orange-300 bg-orange-50/60",
  },
  {
    status: "in_progress",
    title: "In Progress",
    description: "Being actively worked",
    accent: "border-blue-300 bg-blue-50/60",
  },
  {
    status: "waiting_on_customer",
    title: "Waiting",
    description: "Needs customer response",
    accent: "border-amber-300 bg-amber-50/60",
  },
  {
    status: "resolved",
    title: "Resolved",
    description: "Fixed, pending closure",
    accent: "border-green-300 bg-green-50/60",
  },
  {
    status: "closed",
    title: "Closed",
    description: "Finalized tickets",
    accent: "border-slate-300 bg-slate-50/60",
  },
];

function getStatusMeta(status: SupportTicket["status"]) {
  return BOARD_STATUSES.find((item) => item.status === status) ?? BOARD_STATUSES[0];
}

function DraggableTicketCard({
  ticket,
  onEdit,
}: {
  ticket: SupportTicket;
  onEdit: (ticket: SupportTicket) => void;
}) {
  const { attributes, listeners, setNodeRef, transform, isDragging } =
    useDraggable({
      id: ticket.id,
      data: { status: ticket.status },
    });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition: "transform 180ms ease, box-shadow 180ms ease, opacity 180ms ease",
    opacity: isDragging ? 0.55 : 1,
  };

  return (
    <Card
      ref={setNodeRef}
      style={style}
      className={`border shadow-sm ${isDragging ? "ring-2 ring-primary/40 shadow-lg" : ""}`}
    >
      <CardHeader className="space-y-3 p-4 pb-2">
        <div className="flex items-start justify-between gap-2">
          <div className="min-w-0">
            <div className="font-mono text-[11px] text-muted-foreground">
              {ticket.ticket_number}
            </div>
            <CardTitle className="text-sm leading-5">{ticket.subject}</CardTitle>
          </div>
          <div className="flex items-center gap-1 shrink-0">
            <Button variant="ghost" size="icon" className="h-8 w-8" onClick={() => onEdit(ticket)}>
              <Ticket className="h-4 w-4" />
            </Button>
            <Button
              variant="ghost"
              size="icon"
              className="h-8 w-8 cursor-grab active:cursor-grabbing"
              {...attributes}
              {...listeners}
            >
              <GripVertical className="h-4 w-4" />
            </Button>
          </div>
        </div>
        <div className="flex flex-wrap items-center gap-2">
          {getStatusBadge(ticket.status)}
          {getPriorityBadge(ticket.priority)}
          {getSourceBadge(ticket.source)}
        </div>
      </CardHeader>
      <CardContent className="space-y-3 p-4 pt-0">
        <p className="line-clamp-3 text-sm text-muted-foreground">
          {ticket.description}
        </p>
        <div className="space-y-2 text-xs text-muted-foreground">
          <div className="flex items-center gap-2">
            <Building2 className="h-3.5 w-3.5" />
            <span className="truncate">
              {ticket.organization?.name || ticket.organization_id || "Unlinked"}
            </span>
          </div>
          <div className="flex items-center gap-2">
            <User className="h-3.5 w-3.5" />
            <span className="truncate">
              {ticket.user?.email || ticket.user_id || "No user"}
            </span>
          </div>
          <div className="flex items-center gap-2">
            <Calendar className="h-3.5 w-3.5" />
            <span>{format(new Date(ticket.created_at), "MMM dd, yyyy")}</span>
          </div>
        </div>
        <div className="flex items-center justify-between gap-3 rounded-md bg-muted/40 px-3 py-2 text-xs">
          <span className="truncate">
            {ticket.assigned_to_admin?.name ||
              ticket.assigned_to_admin?.email ||
              "Unassigned"}
          </span>
          {ticket.external_reference ? (
            <span className="font-mono text-[11px] text-muted-foreground truncate">
              {ticket.external_reference}
            </span>
          ) : null}
        </div>
      </CardContent>
    </Card>
  );
}

function TicketCardPreview({ ticket }: { ticket: SupportTicket }) {
  return (
    <Card className="border shadow-xl">
      <CardHeader className="space-y-3 p-4 pb-2">
        <div className="min-w-0">
          <div className="font-mono text-[11px] text-muted-foreground">
            {ticket.ticket_number}
          </div>
          <CardTitle className="text-sm leading-5">{ticket.subject}</CardTitle>
        </div>
        <div className="flex flex-wrap items-center gap-2">
          {getStatusBadge(ticket.status)}
          {getPriorityBadge(ticket.priority)}
          {getSourceBadge(ticket.source)}
        </div>
      </CardHeader>
      <CardContent className="space-y-3 p-4 pt-0">
        <p className="line-clamp-3 text-sm text-muted-foreground">
          {ticket.description}
        </p>
        <div className="space-y-2 text-xs text-muted-foreground">
          <div className="flex items-center gap-2">
            <Building2 className="h-3.5 w-3.5" />
            <span className="truncate">
              {ticket.organization?.name || ticket.organization_id || "Unlinked"}
            </span>
          </div>
          <div className="flex items-center gap-2">
            <User className="h-3.5 w-3.5" />
            <span className="truncate">
              {ticket.user?.email || ticket.user_id || "No user"}
            </span>
          </div>
          <div className="flex items-center gap-2">
            <Calendar className="h-3.5 w-3.5" />
            <span>{format(new Date(ticket.created_at), "MMM dd, yyyy")}</span>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

function TicketColumn({
  status,
  tickets,
  onEdit,
}: {
  status: SupportTicket["status"];
  tickets: SupportTicket[];
  onEdit: (ticket: SupportTicket) => void;
}) {
  const { setNodeRef, isOver } = useDroppable({
    id: status,
    data: { status },
  });
  const meta = getStatusMeta(status);

  return (
    <div
      ref={setNodeRef}
      className={`flex min-h-[24rem] flex-col rounded-2xl border p-3 transition-colors ${meta.accent} ${isOver ? "ring-2 ring-primary/30" : ""}`}
    >
      <div className="mb-3 flex items-center justify-between gap-3">
        <div>
          <div className="flex items-center gap-2">
            <h3 className="font-semibold">{meta.title}</h3>
            <Badge variant="secondary">{tickets.length}</Badge>
          </div>
          <p className="text-xs text-muted-foreground">{meta.description}</p>
        </div>
      </div>
      <div className="flex-1 space-y-3">
        {tickets.length === 0 ? (
          <div className="flex h-40 items-center justify-center rounded-xl border border-dashed bg-background/50 text-sm text-muted-foreground">
            Drop tickets here
          </div>
        ) : (
          tickets.map((ticket) => (
            <DraggableTicketCard key={ticket.id} ticket={ticket} onEdit={onEdit} />
          ))
        )}
      </div>
    </div>
  );
}

export default function SupportTicketsPage() {
  const [filters, setFilters] = useState<SupportTicketFilters>({
    page: 1,
    limit: 100,
    priority: "all",
    source: "all",
  });
  const [searchTerm, setSearchTerm] = useState("");
  const [showCreateDialog, setShowCreateDialog] = useState(false);
  const [showEditDialog, setShowEditDialog] = useState(false);
  const [selectedTicket, setSelectedTicket] = useState<SupportTicket | null>(
    null,
  );
  const [activeTicketId, setActiveTicketId] = useState<string | null>(null);
  const [boardTickets, setBoardTickets] = useState<SupportTicket[]>([]);
  const [createForm, setCreateForm] =
    useState<CreateFormState>(defaultCreateForm);
  const [editForm, setEditForm] = useState<EditFormState | null>(null);
  const rollbackTicketsRef = useRef<SupportTicket[] | null>(null);

  const {
    data: ticketData,
    isLoading,
    refetch: refetchTickets,
    isRefetching,
  } = useSupportTickets(filters);
  const { data: stats } = useSupportTicketStats();
  const { data: adminUsers = [] } = useAdminUsers({ is_active: true });

  const createMutation = useCreateSupportTicket();
  const updateMutation = useUpdateSupportTicket();
  const sensors = useSensors(
    useSensor(PointerSensor, {
      activationConstraint: { distance: 8 },
    }),
  );

  const tickets = ticketData?.tickets ?? [];
  useEffect(() => {
    setBoardTickets(tickets);
  }, [tickets]);

  const ticketGroups = useMemo(() => {
    const groups: Record<SupportTicket["status"], SupportTicket[]> = {
      open: [],
      in_progress: [],
      waiting_on_customer: [],
      resolved: [],
      closed: [],
    };

    for (const ticket of boardTickets) {
      groups[ticket.status]?.push(ticket);
    }

    return groups;
  }, [boardTickets]);
  const activeTicket =
    boardTickets.find((ticket) => ticket.id === activeTicketId) ?? null;

  useEffect(() => {
    const timer = setTimeout(() => {
      setFilters((prev) => ({
        ...prev,
        search: searchTerm || undefined,
        page: 1,
      }));
    }, 300);

    return () => clearTimeout(timer);
  }, [searchTerm]);

  const handleRefresh = () => {
    refetchTickets();
  };

  const handleFiltersReset = () => {
    setSearchTerm("");
    setFilters({
      page: 1,
      limit: 100,
      priority: "all",
      source: "all",
    });
  };

  const handleDragStart = (event: DragStartEvent) => {
    setActiveTicketId(String(event.active.id));
  };

  const handleDragEnd = (event: DragEndEvent) => {
    setActiveTicketId(null);

    const ticketId = String(event.active.id);
    const destination = String(event.over?.id ?? "");
    const ticket = boardTickets.find((item) => item.id === ticketId);

    if (!ticket) return;
    if (!BOARD_STATUSES.some((item) => item.status === destination)) return;
    if (ticket.status === destination) return;

    rollbackTicketsRef.current = boardTickets;
    setBoardTickets((current) =>
      current.map((item) =>
        item.id === ticket.id
          ? { ...item, status: destination as SupportTicket["status"] }
          : item,
      ),
    );

    updateMutation.mutate(
      {
        id: ticket.id,
        request: { status: destination as UpdateSupportTicketRequest["status"] },
      },
      {
        onSuccess: (result) => {
          if (result.success) {
            toast.success(
              `Ticket moved to ${getStatusMeta(destination as SupportTicket["status"]).title}`,
            );
          } else {
            if (rollbackTicketsRef.current) {
              setBoardTickets(rollbackTicketsRef.current);
            }
            toast.error(result.message || "Failed to update support ticket");
          }
          rollbackTicketsRef.current = null;
        },
        onError: () => {
          if (rollbackTicketsRef.current) {
            setBoardTickets(rollbackTicketsRef.current);
          }
          rollbackTicketsRef.current = null;
          toast.error("Failed to update support ticket");
        },
      },
    );
  };

  const handleCreateSubmit = async () => {
    const subject = createForm.subject.trim();
    const description = createForm.description.trim();

    if (!subject || !description) {
      toast.error("Subject and description are required");
      return;
    }

    const payload: CreateSupportTicketRequest = {
      subject,
      description,
      category: createForm.category?.trim() || "general",
      priority: createForm.priority || "medium",
      organization_id: createForm.organization_id?.trim() || undefined,
      user_id: createForm.user_id?.trim() || undefined,
      assigned_to_admin_id:
        createForm.assigned_to_admin_id?.trim() || undefined,
      external_reference: createForm.external_reference?.trim() || undefined,
      internal_notes: createForm.internal_notes?.trim() || undefined,
    };

    createMutation.mutate(payload, {
      onSuccess: (result) => {
        if (result.success) {
          toast.success("Support ticket created");
          setCreateForm(defaultCreateForm);
          setShowCreateDialog(false);
        } else {
          toast.error(result.message || "Failed to create support ticket");
        }
      },
    });
  };

  const openEditDialog = (ticket: SupportTicket) => {
    setSelectedTicket(ticket);
    setEditForm({
      id: ticket.id,
      organization_id: ticket.organization_id ?? undefined,
      user_id: ticket.user_id ?? undefined,
      assigned_to_admin_id: ticket.assigned_to_admin_id ?? undefined,
      subject: ticket.subject,
      description: ticket.description,
      category: ticket.category,
      priority: ticket.priority,
      status: ticket.status,
      external_reference: ticket.external_reference ?? undefined,
      internal_notes: ticket.internal_notes ?? undefined,
      resolution_summary: ticket.resolution_summary ?? undefined,
    });
    setShowEditDialog(true);
  };

  const handleUpdateSubmit = async () => {
    if (!editForm) return;

    const payload: UpdateSupportTicketRequest = {
      organization_id: editForm.organization_id,
      user_id: editForm.user_id,
      assigned_to_admin_id: editForm.assigned_to_admin_id,
      subject: editForm.subject?.trim() || undefined,
      description: editForm.description?.trim() || undefined,
      category: editForm.category?.trim() || undefined,
      priority: editForm.priority,
      status: editForm.status,
      external_reference: editForm.external_reference?.trim() || undefined,
      internal_notes: editForm.internal_notes?.trim() || undefined,
      resolution_summary: editForm.resolution_summary?.trim() || undefined,
    };

    updateMutation.mutate(
      { id: editForm.id, request: payload },
      {
        onSuccess: (result) => {
          if (result.success) {
            toast.success("Support ticket updated");
            setShowEditDialog(false);
            setSelectedTicket(null);
            setEditForm(null);
          } else {
            toast.error(result.message || "Failed to update support ticket");
          }
        },
      },
    );
  };

  const statsData = stats ?? {
    total_tickets: 0,
    open_tickets: 0,
    in_progress_tickets: 0,
    waiting_tickets: 0,
    resolved_tickets: 0,
    closed_tickets: 0,
    manual_tickets: 0,
    user_app_tickets: 0,
    email_tickets: 0,
    overdue_tickets: 0,
  };

  return (
    <div className="space-y-6">
      <div className="flex items-start justify-between gap-4">
        <div>
          <h1 className="text-3xl font-bold tracking-tight flex items-center gap-2">
            <Ticket className="h-8 w-8" />
            Tickets
          </h1>
          <p className="text-muted-foreground">
            Manual support queue for customer issues. Future user-app tickets can
            land here without changing the support workflow.
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Button variant="outline" onClick={handleRefresh} disabled={isRefetching}>
            <RefreshCw className={`mr-2 h-4 w-4 ${isRefetching ? "animate-spin" : ""}`} />
            Refresh
          </Button>
          <Button onClick={() => setShowCreateDialog(true)}>
            <Plus className="mr-2 h-4 w-4" />
            Create Ticket
          </Button>
        </div>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Tickets</CardTitle>
            <Ticket className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{statsData.total_tickets}</div>
            <p className="text-xs text-muted-foreground">
              {statsData.manual_tickets} manual entries
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Open</CardTitle>
            <CircleDot className="h-4 w-4 text-orange-600" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{statsData.open_tickets}</div>
            <p className="text-xs text-muted-foreground">
              {statsData.overdue_tickets} overdue
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">In Progress</CardTitle>
            <Clock3 className="h-4 w-4 text-blue-600" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{statsData.in_progress_tickets}</div>
            <p className="text-xs text-muted-foreground">
              {statsData.waiting_tickets} waiting on customer
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Resolved</CardTitle>
            <CheckCircle2 className="h-4 w-4 text-green-600" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{statsData.resolved_tickets}</div>
            <p className="text-xs text-muted-foreground">
              {statsData.closed_tickets} closed
            </p>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="text-base flex items-center gap-2">
            <Filter className="h-4 w-4" />
            Filters
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex flex-wrap gap-3">
            <Input
              placeholder="Search ticket number, subject, or reference"
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="w-[320px]"
            />
            <SelectField
              value={filters.priority ?? "all"}
              onValueChange={(value) =>
                setFilters((prev) => ({
                  ...prev,
                  priority: value === "all" ? "all" : value,
                  page: 1,
                }))
              }
              options={priorityOptions}
              classNames={{ wrapper: "w-44" }}
            />
            <SelectField
              value={filters.source ?? "all"}
              onValueChange={(value) =>
                setFilters((prev) => ({
                  ...prev,
                  source: value === "all" ? "all" : value,
                  page: 1,
                }))
              }
              options={sourceOptions}
              classNames={{ wrapper: "w-40" }}
            />
            <Button variant="outline" onClick={handleFiltersReset}>
              Reset
            </Button>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Support Tickets Board</CardTitle>
          <CardDescription>
            Drag cards between columns to update ticket status. Showing{" "}
            {tickets.length.toLocaleString()} tickets in the current filter set.
          </CardDescription>
        </CardHeader>
        <CardContent>
          {isLoading && tickets.length === 0 ? (
            <div className="grid gap-4 lg:grid-cols-5">
              {BOARD_STATUSES.map((status) => (
                <div
                  key={status.status}
                  className="min-h-[24rem] rounded-2xl border bg-muted/20 p-3"
                >
                  <div className="mb-3 h-5 w-24 rounded bg-muted animate-pulse" />
                  <div className="space-y-3">
                    {[...Array(3)].map((_, index) => (
                      <div
                        key={`${status.status}-skeleton-${index}`}
                        className="h-32 rounded-xl border bg-background animate-pulse"
                      />
                    ))}
                  </div>
                </div>
              ))}
            </div>
          ) : tickets.length === 0 ? (
            <div className="text-center py-12">
              <AlertTriangle className="h-10 w-10 text-muted-foreground mx-auto mb-3" />
              <p className="text-muted-foreground">No tickets found</p>
            </div>
          ) : (
            <DndContext
              sensors={sensors}
              collisionDetection={closestCenter}
              onDragStart={handleDragStart}
              onDragEnd={handleDragEnd}
            >
              <div className="overflow-x-auto pb-2">
                <div className="grid min-w-[1200px] gap-4 lg:grid-cols-5">
                  {BOARD_STATUSES.map((status) => (
                    <TicketColumn
                      key={status.status}
                      status={status.status}
                      tickets={ticketGroups[status.status]}
                      onEdit={openEditDialog}
                    />
                  ))}
                </div>
              </div>
              <DragOverlay adjustScale={false}>
                {activeTicket ? <TicketCardPreview ticket={activeTicket} /> : null}
              </DragOverlay>
              {activeTicket ? (
                <div className="mt-4 rounded-lg border bg-muted/30 px-4 py-3 text-sm text-muted-foreground">
                  Dragging {activeTicket.ticket_number} • Drop it on a new status column to update it.
                </div>
              ) : null}
            </DndContext>
          )}
        </CardContent>
      </Card>

      <Dialog open={showCreateDialog} onOpenChange={setShowCreateDialog}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>Create Ticket</DialogTitle>
            <DialogDescription>
              Create a manual support ticket for a customer issue.
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4">
            <Input
              placeholder="Subject"
              value={createForm.subject}
              onChange={(e) =>
                setCreateForm((prev) => ({ ...prev, subject: e.target.value }))
              }
            />
            <div className="grid gap-4 md:grid-cols-2">
              <SelectField
                value={createForm.category ?? "general"}
                onValueChange={(value) =>
                  setCreateForm((prev) => ({ ...prev, category: value }))
                }
                options={categoryOptions}
              />
              <SelectField
                value={createForm.priority ?? "medium"}
                onValueChange={(value) =>
                  setCreateForm((prev) => ({
                    ...prev,
                    priority: value as CreateSupportTicketRequest["priority"],
                  }))
                }
                options={priorityOptions.filter((option) => option.value !== "all")}
              />
            </div>
            <div className="grid gap-4 md:grid-cols-2">
              <Input
                placeholder="Organization ID (optional)"
                value={createForm.organization_id ?? ""}
                onChange={(e) =>
                  setCreateForm((prev) => ({
                    ...prev,
                    organization_id: e.target.value,
                  }))
                }
              />
              <Input
                placeholder="User ID (optional)"
                value={createForm.user_id ?? ""}
                onChange={(e) =>
                  setCreateForm((prev) => ({ ...prev, user_id: e.target.value }))
                }
              />
            </div>
            <SelectField
              label="Assign To"
              placeholder="Unassigned"
              value={createForm.assigned_to_admin_id ?? ""}
              onValueChange={(value) =>
                setCreateForm((prev) => ({
                  ...prev,
                  assigned_to_admin_id: value,
                }))
              }
              options={[
                { value: "", label: "Unassigned" },
                ...adminUsers.map((user) => ({
                  value: user.id,
                  label: `${user.full_name} (${user.email})`,
                })),
              ]}
            />
            <Textarea
              placeholder="Describe the issue"
              value={createForm.description}
              onChange={(e) =>
                setCreateForm((prev) => ({
                  ...prev,
                  description: e.target.value,
                }))
              }
              rows={6}
            />
            <Textarea
              placeholder="Internal notes (optional)"
              value={createForm.internal_notes ?? ""}
              onChange={(e) =>
                setCreateForm((prev) => ({
                  ...prev,
                  internal_notes: e.target.value,
                }))
              }
              rows={3}
            />
            <Input
              placeholder="External reference (optional)"
              value={createForm.external_reference ?? ""}
              onChange={(e) =>
                setCreateForm((prev) => ({
                  ...prev,
                  external_reference: e.target.value,
                }))
              }
            />
            <div className="flex justify-end gap-2">
              <Button variant="outline" onClick={() => setShowCreateDialog(false)}>
                Cancel
              </Button>
              <Button
                onClick={handleCreateSubmit}
                disabled={createMutation.isPending}
              >
                {createMutation.isPending ? "Creating..." : "Create Ticket"}
              </Button>
            </div>
          </div>
        </DialogContent>
      </Dialog>

      <Dialog open={showEditDialog} onOpenChange={setShowEditDialog}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>Ticket Details</DialogTitle>
            <DialogDescription>
              Update ticket status, assignment, and internal notes.
            </DialogDescription>
          </DialogHeader>

          {selectedTicket && editForm && (
            <div className="space-y-4">
              <div className="rounded-lg border p-3">
                <div className="text-xs font-mono text-muted-foreground">
                  {selectedTicket.ticket_number}
                </div>
                <div className="text-sm font-medium">{selectedTicket.subject}</div>
                <div className="text-xs text-muted-foreground">
                  Created {format(new Date(selectedTicket.created_at), "PPpp")}
                </div>
              </div>
              <div className="grid gap-4 md:grid-cols-2">
                <SelectField
                  label="Status"
                  value={editForm.status ?? selectedTicket.status}
                  onValueChange={(value) =>
                    setEditForm((prev) =>
                      prev ? { ...prev, status: value as EditFormState["status"] } : prev,
                    )
                  }
                  options={statusOptions.filter((option) => option.value !== "all")}
                />
                <SelectField
                  label="Priority"
                  value={editForm.priority ?? selectedTicket.priority}
                  onValueChange={(value) =>
                    setEditForm((prev) =>
                      prev ? { ...prev, priority: value as EditFormState["priority"] } : prev,
                    )
                  }
                  options={priorityOptions.filter((option) => option.value !== "all")}
                />
              </div>
              <SelectField
                label="Assigned To"
                placeholder="Unassigned"
                value={editForm.assigned_to_admin_id ?? ""}
                onValueChange={(value) =>
                  setEditForm((prev) =>
                    prev ? { ...prev, assigned_to_admin_id: value } : prev,
                  )
                }
                options={[
                  { value: "", label: "Unassigned" },
                  ...adminUsers.map((user) => ({
                    value: user.id,
                    label: `${user.full_name} (${user.email})`,
                  })),
                ]}
              />
              <div className="space-y-2">
                <Label>Description</Label>
                <Textarea
                  value={editForm.description ?? ""}
                  onChange={(e) =>
                    setEditForm((prev) =>
                      prev ? { ...prev, description: e.target.value } : prev,
                    )
                  }
                  rows={5}
                />
              </div>
              <div className="space-y-2">
                <Label>Internal Notes</Label>
                <Textarea
                  value={editForm.internal_notes ?? ""}
                  onChange={(e) =>
                    setEditForm((prev) =>
                      prev ? { ...prev, internal_notes: e.target.value } : prev,
                    )
                  }
                  rows={3}
                />
              </div>
              <div className="space-y-2">
                <Label>Resolution Summary</Label>
                <Textarea
                  value={editForm.resolution_summary ?? ""}
                  onChange={(e) =>
                    setEditForm((prev) =>
                      prev
                        ? { ...prev, resolution_summary: e.target.value }
                        : prev,
                    )
                  }
                  rows={3}
                />
              </div>
              <div className="flex justify-end gap-2">
                <Button variant="outline" onClick={() => setShowEditDialog(false)}>
                  Cancel
                </Button>
                <Button
                  onClick={handleUpdateSubmit}
                  disabled={updateMutation.isPending}
                >
                  {updateMutation.isPending ? "Saving..." : "Save Changes"}
                </Button>
              </div>
            </div>
          )}
        </DialogContent>
      </Dialog>
    </div>
  );
}
