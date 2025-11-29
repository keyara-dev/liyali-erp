import { fetchCurrencies } from "@/app/_actions/currencies";
import { fetchEventHosts } from "@/app/_actions/event-hosts";
import { fetchWhatsAppConfig } from "@/app/_actions/whatsapp";
import { QUERY_KEYS } from "@/lib/constants";
import { useQuery } from "@tanstack/react-query";

export const useSellerProfiles = () =>
  useQuery({
    queryKey: [QUERY_KEYS.SELLERS],
    queryFn: async () => await fetchEventHosts(),
    staleTime: Infinity,
  });

export const useCurrencies = () =>
  useQuery({
    queryKey: [QUERY_KEYS.CURRENCIES],
    queryFn: async () => await fetchCurrencies(),
    staleTime: Infinity,
  });

export const useWhatsAppConfig = () =>
  useQuery({
    queryKey: [QUERY_KEYS.WHATSAPP_CONFIG],
    queryFn: async () => await fetchWhatsAppConfig(),
    staleTime: 5 * 60 * 1000, // 5 minutes
  });
