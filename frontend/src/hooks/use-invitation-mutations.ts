"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import {
  sendInvitation,
  cancelInvitation,
  resendInvitation,
  acceptInvitation,
  declineInvitation,
  listOrgInvitations,
  getMyPendingInvitations,
  type SendInvitationRequest,
} from "@/app/_actions/invitation-actions";

const INVITATIONS_KEY = ["invitations"];
const MY_INVITATIONS_KEY = ["invitations", "pending"];

// ─── Admin-side hooks ─────────────────────────────────────────────────────────

export function useOrgInvitations() {
  return useQuery({
    queryKey: INVITATIONS_KEY,
    queryFn: async () => {
      const res = await listOrgInvitations();
      if (!res.success) throw new Error(res.message);
      return res.data ?? [];
    },
  });
}

export function useSendInvitation(onSuccess?: () => void) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (data: SendInvitationRequest) => {
      const res = await sendInvitation(data);
      if (!res.success) throw new Error(res.message || "Failed to send invitation");
      return res;
    },
    onSuccess: () => {
      toast.success("Invitation sent successfully");
      queryClient.invalidateQueries({ queryKey: INVITATIONS_KEY });
      onSuccess?.();
    },
    onError: (err: Error) => {
      toast.error(err.message || "Failed to send invitation");
    },
  });
}

export function useCancelInvitation(onSuccess?: () => void) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (id: string) => {
      const res = await cancelInvitation(id);
      if (!res.success) throw new Error(res.message || "Failed to cancel invitation");
      return res;
    },
    onSuccess: () => {
      toast.success("Invitation cancelled");
      queryClient.invalidateQueries({ queryKey: INVITATIONS_KEY });
      onSuccess?.();
    },
    onError: (err: Error) => {
      toast.error(err.message || "Failed to cancel invitation");
    },
  });
}

export function useResendInvitation(onSuccess?: () => void) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (id: string) => {
      const res = await resendInvitation(id);
      if (!res.success) throw new Error(res.message || "Failed to resend invitation");
      return res;
    },
    onSuccess: () => {
      toast.success("Invitation resent");
      queryClient.invalidateQueries({ queryKey: INVITATIONS_KEY });
      onSuccess?.();
    },
    onError: (err: Error) => {
      toast.error(err.message || "Failed to resend invitation");
    },
  });
}

// ─── Invitee-facing hooks ─────────────────────────────────────────────────────

export function useMyPendingInvitations() {
  return useQuery({
    queryKey: MY_INVITATIONS_KEY,
    queryFn: async () => {
      const res = await getMyPendingInvitations();
      if (!res.success) throw new Error(res.message);
      return res.data ?? [];
    },
  });
}

export function useAcceptInvitation(onSuccess?: () => void) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (token: string) => {
      const res = await acceptInvitation(token);
      if (!res.success) throw new Error(res.message || "Failed to accept invitation");
      return res;
    },
    onSuccess: () => {
      toast.success("Invitation accepted — welcome to the organization!");
      queryClient.invalidateQueries({ queryKey: MY_INVITATIONS_KEY });
      queryClient.invalidateQueries({ queryKey: ["notifications"] });
      onSuccess?.();
    },
    onError: (err: Error) => {
      toast.error(err.message || "Failed to accept invitation");
    },
  });
}

export function useDeclineInvitation(onSuccess?: () => void) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (token: string) => {
      const res = await declineInvitation(token);
      if (!res.success) throw new Error(res.message || "Failed to decline invitation");
      return res;
    },
    onSuccess: () => {
      toast.success("Invitation declined");
      queryClient.invalidateQueries({ queryKey: MY_INVITATIONS_KEY });
      queryClient.invalidateQueries({ queryKey: ["notifications"] });
      onSuccess?.();
    },
    onError: (err: Error) => {
      toast.error(err.message || "Failed to decline invitation");
    },
  });
}
