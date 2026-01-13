import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import type {
  CreateTravelStoryRequest,
  DetectTripsRequest,
  DetectTripsResponse,
  GenerateContentRequest,
  GenerateContentResponse,
  ListTravelStoriesResponse,
  TravelStory,
} from "@/types/travel-story";

const API_BASE = "/api/v1/travelStories";

// Query keys factory
export const travelStoryKeys = {
  all: ["travelStories"] as const,
  lists: () => [...travelStoryKeys.all, "list"] as const,
  list: (filters?: Record<string, unknown>) => [...travelStoryKeys.lists(), filters] as const,
  details: () => [...travelStoryKeys.all, "detail"] as const,
  detail: (uid: string) => [...travelStoryKeys.details(), uid] as const,
  trips: () => [...travelStoryKeys.all, "trips"] as const,
};

// Fetch helper with error handling
async function fetchAPI<T>(url: string, options?: RequestInit): Promise<T> {
  const response = await fetch(url, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...options?.headers,
    },
  });

  if (!response.ok) {
    const error = await response.json().catch(() => ({ message: response.statusText }));
    throw new Error(error.message || "API request failed");
  }

  // Handle 204 No Content
  if (response.status === 204) {
    return {} as T;
  }

  return response.json();
}

// List travel stories
export function useTravelStories() {
  return useQuery({
    queryKey: travelStoryKeys.list(),
    queryFn: async () => {
      const response = await fetchAPI<ListTravelStoriesResponse>(API_BASE);
      return response.travelStories;
    },
    staleTime: 1000 * 60, // 1 minute
  });
}

// Get single travel story
export function useTravelStory(uid: string, options?: { enabled?: boolean }) {
  return useQuery({
    queryKey: travelStoryKeys.detail(uid),
    queryFn: async () => {
      return fetchAPI<TravelStory>(`${API_BASE}/${uid}`);
    },
    enabled: options?.enabled ?? !!uid,
    staleTime: 1000 * 60,
  });
}

// Create travel story
export function useCreateTravelStory() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (request: CreateTravelStoryRequest) => {
      return fetchAPI<TravelStory>(API_BASE, {
        method: "POST",
        body: JSON.stringify(request),
      });
    },
    onSuccess: (newStory) => {
      queryClient.invalidateQueries({ queryKey: travelStoryKeys.lists() });
      queryClient.setQueryData(travelStoryKeys.detail(newStory.name.replace("travelStories/", "")), newStory);
    },
  });
}

// Update travel story
export function useUpdateTravelStory() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ uid, updates }: { uid: string; updates: Partial<TravelStory> }) => {
      return fetchAPI<TravelStory>(`${API_BASE}/${uid}`, {
        method: "PATCH",
        body: JSON.stringify(updates),
      });
    },
    onSuccess: (updatedStory, { uid }) => {
      queryClient.setQueryData(travelStoryKeys.detail(uid), updatedStory);
      queryClient.invalidateQueries({ queryKey: travelStoryKeys.lists() });
    },
  });
}

// Delete travel story
export function useDeleteTravelStory() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (uid: string) => {
      await fetchAPI(`${API_BASE}/${uid}`, { method: "DELETE" });
      return uid;
    },
    onSuccess: (uid) => {
      queryClient.removeQueries({ queryKey: travelStoryKeys.detail(uid) });
      queryClient.invalidateQueries({ queryKey: travelStoryKeys.lists() });
    },
  });
}

// Detect trips from memos
export function useDetectTrips() {
  return useMutation({
    mutationFn: async (request: DetectTripsRequest) => {
      return fetchAPI<DetectTripsResponse>(`${API_BASE}/detectTrips`, {
        method: "POST",
        body: JSON.stringify(request),
      });
    },
  });
}

// Generate content from memos
export function useGenerateContent() {
  return useMutation({
    mutationFn: async (request: GenerateContentRequest) => {
      return fetchAPI<GenerateContentResponse>(`${API_BASE}/generateContent`, {
        method: "POST",
        body: JSON.stringify(request),
      });
    },
  });
}

// Get export URL for a travel story
export function getTravelStoryExportUrl(uid: string, format: "html" | "pdf" = "html"): string {
  return `${API_BASE}/${uid}/export?format=${format}`;
}

// Get shareable view URL for a travel story
export function getTravelStoryViewUrl(uid: string): string {
  return `${API_BASE}/${uid}/view`;
}
