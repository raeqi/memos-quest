import { useParams } from "react-router-dom";
import { CalendarIcon, DownloadIcon, GlobeIcon, Loader2Icon, MapPinIcon, ShareIcon } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useTravelStory, getTravelStoryExportUrl, getTravelStoryViewUrl } from "@/hooks/useTravelStoryQueries";
import { cn } from "@/lib/utils";
import { useTranslate } from "@/utils/i18n";

const TravelStoryViewer = () => {
  const t = useTranslate();
  const { uid } = useParams<{ uid: string }>();
  const { data: story, isLoading, error } = useTravelStory(uid || "");

  const handleShare = async () => {
    if (!uid) return;
    const url = window.location.origin + getTravelStoryViewUrl(uid);
    try {
      if (navigator.share) {
        await navigator.share({
          title: story?.title,
          text: story?.description,
          url,
        });
      } else {
        await navigator.clipboard.writeText(url);
        alert(t("message.copied", "Copied to clipboard!"));
      }
    } catch (err) {
      console.error("Share failed:", err);
    }
  };

  const handleExport = () => {
    if (!uid) return;
    window.open(getTravelStoryExportUrl(uid, "html"), "_blank");
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <Loader2Icon className="w-8 h-8 animate-spin text-primary" />
      </div>
    );
  }

  if (error || !story) {
    return (
      <div className="flex flex-col items-center justify-center min-h-[400px] text-muted-foreground">
        <p>{t("message.memo-not-found", "Story not found")}</p>
      </div>
    );
  }

  const formatDate = (timestamp: number) => {
    return new Date(timestamp * 1000).toLocaleDateString(undefined, {
      year: "numeric",
      month: "long",
      day: "numeric",
    });
  };

  const themeColors = {
    ocean: { accent: "bg-blue-600", bg: "bg-blue-50", text: "text-blue-900" },
    sunset: { accent: "bg-orange-600", bg: "bg-orange-50", text: "text-orange-900" },
    forest: { accent: "bg-green-600", bg: "bg-green-50", text: "text-green-900" },
    minimal: { accent: "bg-gray-800", bg: "bg-gray-50", text: "text-gray-900" },
  };

  const theme = themeColors[(story.theme?.colorScheme as keyof typeof themeColors) || "ocean"];

  return (
    <div className={cn("min-h-screen", theme.bg)}>
      {/* Header */}
      <header className={cn("relative py-16 px-4", theme.accent, "text-white")}>
        <div className="max-w-4xl mx-auto text-center">
          <h1 className="text-4xl md:text-5xl font-bold mb-4">{story.title}</h1>
          {story.description && <p className="text-xl opacity-90 mb-6">{story.description}</p>}
          <div className="flex flex-wrap items-center justify-center gap-4 text-sm opacity-80">
            {story.destination && (
              <span className="flex items-center gap-1">
                <MapPinIcon className="w-4 h-4" />
                {story.destination}
              </span>
            )}
            {story.startDate && story.endDate && (
              <span className="flex items-center gap-1">
                <CalendarIcon className="w-4 h-4" />
                {formatDate(story.startDate)} - {formatDate(story.endDate)}
              </span>
            )}
          </div>
        </div>

        {/* Action buttons */}
        <div className="absolute top-4 right-4 flex gap-2">
          <Button variant="secondary" size="sm" onClick={handleShare}>
            <ShareIcon className="w-4 h-4 mr-1" />
            {t("common.share", "Share")}
          </Button>
          <Button variant="secondary" size="sm" onClick={handleExport}>
            <DownloadIcon className="w-4 h-4 mr-1" />
            Export
          </Button>
        </div>
      </header>

      {/* Content */}
      <main className="max-w-4xl mx-auto px-4 py-12">
        {/* Cover image */}
        {story.coverImage && (
          <div className="mb-12 -mt-20 relative">
            <img
              src={story.coverImage}
              alt={story.title}
              className="w-full h-64 md:h-96 object-cover rounded-xl shadow-lg"
            />
          </div>
        )}

        {/* Tags */}
        {story.tags && story.tags.length > 0 && (
          <div className="flex flex-wrap gap-2 mb-8">
            {story.tags.map((tag) => (
              <span
                key={tag}
                className={cn("px-3 py-1 rounded-full text-sm", theme.accent, "text-white opacity-80")}
              >
                #{tag}
              </span>
            ))}
          </div>
        )}

        {/* Sections */}
        <div className="space-y-12">
          {story.sections.map((section, index) => (
            <article key={index} className="bg-white rounded-xl shadow-sm p-6 md:p-8">
              {section.title && <h2 className={cn("text-2xl font-semibold mb-4", theme.text)}>{section.title}</h2>}

              <div className="flex flex-wrap gap-4 text-sm text-muted-foreground mb-4">
                {section.date && (
                  <span className="flex items-center gap-1">
                    <CalendarIcon className="w-4 h-4" />
                    {formatDate(section.date)}
                  </span>
                )}
                {section.location && (
                  <span className="flex items-center gap-1">
                    <MapPinIcon className="w-4 h-4" />
                    {section.location}
                  </span>
                )}
              </div>

              <div className="prose prose-lg max-w-none whitespace-pre-wrap">{section.content}</div>

              {/* Section images */}
              {section.images && section.images.length > 0 && (
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 mt-6">
                  {section.images
                    .filter((img) => img.selected)
                    .map((image, imgIndex) => (
                      <div key={imgIndex} className="relative aspect-[4/3] overflow-hidden rounded-lg">
                        <img
                          src={image.url}
                          alt={image.altText || `Image ${imgIndex + 1}`}
                          className="w-full h-full object-cover hover:scale-105 transition-transform duration-300"
                          loading="lazy"
                        />
                        {image.caption && (
                          <div className="absolute bottom-0 left-0 right-0 bg-gradient-to-t from-black/60 to-transparent p-3">
                            <p className="text-white text-sm">{image.caption}</p>
                          </div>
                        )}
                      </div>
                    ))}
                </div>
              )}
            </article>
          ))}
        </div>

        {/* Footer */}
        <footer className="mt-16 pt-8 border-t text-center text-muted-foreground">
          <p className="flex items-center justify-center gap-2">
            <GlobeIcon className="w-4 h-4" />
            Created with Memos
          </p>
        </footer>
      </main>
    </div>
  );
};

export default TravelStoryViewer;
