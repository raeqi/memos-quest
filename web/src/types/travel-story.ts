// Travel Story types for the Travel Journal feature

export interface TravelStory {
  name: string;
  creator: string;
  title: string;
  description: string;
  coverImage: string;
  startDate?: number;
  endDate?: number;
  destination: string;
  tags: string[];
  sections: StorySection[];
  theme?: StoryTheme;
  visibility: "PUBLIC" | "PROTECTED" | "PRIVATE";
  createTime: number;
  updateTime: number;
  sourceMemos: string[];
}

export interface StorySection {
  title: string;
  content: string;
  date?: number;
  location: string;
  images: StoryImage[];
  sourceMemo: string;
}

export interface StoryImage {
  url: string;
  caption: string;
  altText: string;
  qualityScore: number;
  selected: boolean;
}

export interface StoryTheme {
  colorScheme: string;
  fontFamily: string;
  layoutStyle: string;
}

export interface DetectedTrip {
  suggestedTitle: string;
  destination: string;
  startDate: number;
  endDate: number;
  tags: string[];
  memoNames: string[];
  confidence: number;
}

export interface CreateTravelStoryRequest {
  title: string;
  description?: string;
  memoNames: string[];
  theme?: StoryTheme;
  visibility?: "PUBLIC" | "PROTECTED" | "PRIVATE";
  autoGenerateContent?: boolean;
}

export interface DetectTripsRequest {
  tags?: string[];
  startDate?: number;
  endDate?: number;
  minMemoCount?: number;
}

export interface GenerateContentRequest {
  memoNames: string[];
  titleContext?: string;
}

export interface GenerateContentResponse {
  sections: StorySection[];
  suggestedCoverImage: string;
  suggestedTitle: string;
  detectedDestination: string;
}

export interface ListTravelStoriesResponse {
  travelStories: TravelStory[];
}

export interface DetectTripsResponse {
  trips: DetectedTrip[];
}

// Theme presets
export const THEME_PRESETS = {
  ocean: {
    colorScheme: "ocean",
    fontFamily: "Georgia, serif",
    layoutStyle: "magazine",
  },
  sunset: {
    colorScheme: "sunset",
    fontFamily: "Palatino, serif",
    layoutStyle: "timeline",
  },
  forest: {
    colorScheme: "forest",
    fontFamily: "Merriweather, serif",
    layoutStyle: "gallery",
  },
  minimal: {
    colorScheme: "minimal",
    fontFamily: "system-ui, sans-serif",
    layoutStyle: "minimal",
  },
} as const;
