import { Link } from "react-router-dom";
import { BookOpenIcon, CalendarIcon, MapPinIcon, PlusIcon, Loader2Icon, EyeIcon, TrashIcon } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useTravelStories, useDeleteTravelStory } from "@/hooks/useTravelStoryQueries";
import { cn } from "@/lib/utils";
import { useTranslate } from "@/utils/i18n";

const TravelStoriesPage = () => {
  const t = useTranslate();
  const { data: stories, isLoading } = useTravelStories();
  const deleteStory = useDeleteTravelStory();

  const formatDate = (timestamp: number) => {
    return new Date(timestamp * 1000).toLocaleDateString(undefined, {
      year: "numeric",
      month: "short",
      day: "numeric",
    });
  };

  const handleDelete = async (uid: string, title: string) => {
    if (confirm(`Are you sure you want to delete "${title}"?`)) {
      await deleteStory.mutateAsync(uid);
    }
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <Loader2Icon className="w-8 h-8 animate-spin text-primary" />
      </div>
    );
  }

  return (
    <div className="w-full max-w-4xl mx-auto py-6">
      <div className="flex items-center justify-between mb-8">
        <div className="flex items-center gap-3">
          <BookOpenIcon className="w-8 h-8 text-primary" />
          <h1 className="text-2xl font-bold">{t("travel-story.stories-title", "Travel Stories")}</h1>
        </div>
        <Link to="/">
          <Button variant="outline" size="sm">
            <PlusIcon className="w-4 h-4 mr-1" />
            {t("travel-story.create-new", "Create New")}
          </Button>
        </Link>
      </div>

      {!stories || stories.length === 0 ? (
        <div className="text-center py-16">
          <BookOpenIcon className="w-16 h-16 mx-auto text-muted-foreground/50 mb-4" />
          <h2 className="text-xl font-medium mb-2">{t("travel-story.no-stories", "No travel stories yet")}</h2>
          <p className="text-muted-foreground mb-6">
            {t("travel-story.no-stories-hint", "Select memos from your journal to create your first travel story.")}
          </p>
          <Link to="/">
            <Button>
              <PlusIcon className="w-4 h-4 mr-2" />
              {t("travel-story.get-started", "Get Started")}
            </Button>
          </Link>
        </div>
      ) : (
        <div className="grid gap-6 md:grid-cols-2">
          {stories.map((story) => {
            const uid = story.name.replace("travelStories/", "");
            const themeColors: Record<string, string> = {
              ocean: "bg-gradient-to-br from-blue-500 to-blue-700",
              sunset: "bg-gradient-to-br from-orange-500 to-red-600",
              forest: "bg-gradient-to-br from-green-500 to-green-700",
              minimal: "bg-gradient-to-br from-gray-600 to-gray-800",
            };
            const bgClass = themeColors[story.theme?.colorScheme || "ocean"];

            return (
              <div
                key={story.name}
                className="group relative bg-card rounded-xl shadow-sm overflow-hidden border hover:shadow-md transition-shadow"
              >
                {/* Cover/Header */}
                <div className={cn("h-32 relative", story.coverImage ? "" : bgClass)}>
                  {story.coverImage ? (
                    <img src={story.coverImage} alt={story.title} className="w-full h-full object-cover" />
                  ) : (
                    <div className="w-full h-full flex items-center justify-center">
                      <BookOpenIcon className="w-12 h-12 text-white/50" />
                    </div>
                  )}
                  <div className="absolute inset-0 bg-gradient-to-t from-black/60 to-transparent" />
                  <div className="absolute bottom-3 left-4 right-4">
                    <h3 className="text-lg font-semibold text-white truncate">{story.title}</h3>
                  </div>
                </div>

                {/* Content */}
                <div className="p-4">
                  {story.description && (
                    <p className="text-sm text-muted-foreground line-clamp-2 mb-3">{story.description}</p>
                  )}

                  <div className="flex flex-wrap gap-3 text-xs text-muted-foreground">
                    {story.destination && (
                      <span className="flex items-center gap-1">
                        <MapPinIcon className="w-3 h-3" />
                        {story.destination}
                      </span>
                    )}
                    {story.startDate && (
                      <span className="flex items-center gap-1">
                        <CalendarIcon className="w-3 h-3" />
                        {formatDate(story.startDate)}
                      </span>
                    )}
                  </div>

                  {/* Tags */}
                  {story.tags && story.tags.length > 0 && (
                    <div className="flex flex-wrap gap-1.5 mt-3">
                      {story.tags.slice(0, 3).map((tag) => (
                        <span key={tag} className="px-2 py-0.5 bg-muted rounded text-xs">
                          #{tag}
                        </span>
                      ))}
                      {story.tags.length > 3 && (
                        <span className="px-2 py-0.5 text-xs text-muted-foreground">+{story.tags.length - 3}</span>
                      )}
                    </div>
                  )}
                </div>

                {/* Actions */}
                <div className="absolute top-2 right-2 flex gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                  <Link to={`/stories/${uid}`}>
                    <Button variant="secondary" size="icon" className="w-8 h-8">
                      <EyeIcon className="w-4 h-4" />
                    </Button>
                  </Link>
                  <Button
                    variant="secondary"
                    size="icon"
                    className="w-8 h-8 hover:bg-destructive hover:text-destructive-foreground"
                    onClick={() => handleDelete(uid, story.title)}
                  >
                    <TrashIcon className="w-4 h-4" />
                  </Button>
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
};

export default TravelStoriesPage;
