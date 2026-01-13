import { useState } from "react";
import { BookOpenIcon, GlobeIcon, ImageIcon, Loader2Icon, SparklesIcon } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useMemoSelection } from "@/contexts/MemoSelectionContext";
import { useCreateTravelStory, useGenerateContent, getTravelStoryViewUrl } from "@/hooks/useTravelStoryQueries";
import { THEME_PRESETS, type StoryTheme } from "@/types/travel-story";
import { useTranslate } from "@/utils/i18n";

interface Props {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

const GenerateTravelStoryDialog = ({ open, onOpenChange }: Props) => {
  const t = useTranslate();
  const { selectedMemos, clearSelection, toggleSelectionMode } = useMemoSelection();
  const createTravelStory = useCreateTravelStory();
  const generateContent = useGenerateContent();

  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [visibility, setVisibility] = useState<"PUBLIC" | "PROTECTED" | "PRIVATE">("PRIVATE");
  const [themePreset, setThemePreset] = useState<keyof typeof THEME_PRESETS>("ocean");
  const [autoGenerate, setAutoGenerate] = useState(true);
  const [isGenerating, setIsGenerating] = useState(false);
  const [createdStoryUrl, setCreatedStoryUrl] = useState<string | null>(null);

  const selectedMemoNames = Array.from(selectedMemos.keys());

  const handleGeneratePreview = async () => {
    if (selectedMemoNames.length === 0) return;

    setIsGenerating(true);
    try {
      const result = await generateContent.mutateAsync({
        memoNames: selectedMemoNames,
        titleContext: title,
      });

      if (result.suggestedTitle && !title) {
        setTitle(result.suggestedTitle);
      }
      if (result.detectedDestination && !description) {
        setDescription(`A journey to ${result.detectedDestination}`);
      }
    } catch (error) {
      console.error("Failed to generate preview:", error);
    } finally {
      setIsGenerating(false);
    }
  };

  const handleCreate = async () => {
    if (!title || selectedMemoNames.length === 0) return;

    setIsGenerating(true);
    try {
      const theme: StoryTheme = THEME_PRESETS[themePreset];

      const story = await createTravelStory.mutateAsync({
        title,
        description,
        memoNames: selectedMemoNames,
        theme,
        visibility,
        autoGenerateContent: autoGenerate,
      });

      // Get the UID from the story name
      const uid = story.name.replace("travelStories/", "");
      setCreatedStoryUrl(getTravelStoryViewUrl(uid));
    } catch (error) {
      console.error("Failed to create travel story:", error);
    } finally {
      setIsGenerating(false);
    }
  };

  const handleClose = () => {
    if (createdStoryUrl) {
      clearSelection();
      toggleSelectionMode();
    }
    setTitle("");
    setDescription("");
    setCreatedStoryUrl(null);
    onOpenChange(false);
  };

  const handleViewStory = () => {
    if (createdStoryUrl) {
      window.open(createdStoryUrl, "_blank");
    }
  };

  if (createdStoryUrl) {
    return (
      <Dialog open={open} onOpenChange={handleClose}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <SparklesIcon className="w-5 h-5 text-green-500" />
              {t("travel-story.created-title", "Story Created!")}
            </DialogTitle>
            <DialogDescription>
              {t("travel-story.created-description", "Your travel story has been created and is ready to share.")}
            </DialogDescription>
          </DialogHeader>

          <div className="flex flex-col gap-4 py-4">
            <div className="p-4 bg-muted rounded-lg">
              <h3 className="font-semibold text-lg">{title}</h3>
              {description && <p className="text-sm text-muted-foreground mt-1">{description}</p>}
              <p className="text-xs text-muted-foreground mt-2">
                {selectedMemoNames.length} {t("travel-story.memos-included", "memos included")}
              </p>
            </div>
          </div>

          <DialogFooter className="flex gap-2">
            <Button variant="outline" onClick={handleClose}>
              {t("common.close", "Close")}
            </Button>
            <Button onClick={handleViewStory}>
              <GlobeIcon className="w-4 h-4 mr-2" />
              {t("travel-story.view-story", "View Story")}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    );
  }

  return (
    <Dialog open={open} onOpenChange={handleClose}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <BookOpenIcon className="w-5 h-5" />
            {t("travel-story.generate-title", "Generate Travel Story")}
          </DialogTitle>
          <DialogDescription>
            {t(
              "travel-story.generate-description",
              "Turn your selected memos into a beautiful shareable travel journal.",
            )}
          </DialogDescription>
        </DialogHeader>

        <div className="flex flex-col gap-4 py-4">
          {/* Selected memos count */}
          <div className="flex items-center gap-2 p-3 bg-muted rounded-lg">
            <ImageIcon className="w-4 h-4 text-muted-foreground" />
            <span className="text-sm">
              <strong>{selectedMemoNames.length}</strong> {t("travel-story.memos-selected", "memos selected")}
            </span>
          </div>

          {/* Title */}
          <div className="flex flex-col gap-2">
            <Label htmlFor="title">{t("common.title", "Title")}</Label>
            <div className="flex gap-2">
              <Input
                id="title"
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                placeholder={t("travel-story.title-placeholder", "My Amazing Trip")}
              />
              <Button
                variant="outline"
                size="icon"
                onClick={handleGeneratePreview}
                disabled={isGenerating || selectedMemoNames.length === 0}
                title={t("travel-story.auto-generate", "Auto-generate from memos")}
              >
                {isGenerating ? <Loader2Icon className="w-4 h-4 animate-spin" /> : <SparklesIcon className="w-4 h-4" />}
              </Button>
            </div>
          </div>

          {/* Description */}
          <div className="flex flex-col gap-2">
            <Label htmlFor="description">{t("common.description", "Description")}</Label>
            <Textarea
              id="description"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder={t("travel-story.description-placeholder", "A brief description of your journey...")}
              rows={2}
            />
          </div>

          {/* Theme */}
          <div className="flex flex-col gap-2">
            <Label>{t("travel-story.theme", "Theme")}</Label>
            <Select value={themePreset} onValueChange={(v) => setThemePreset(v as keyof typeof THEME_PRESETS)}>
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="ocean">{t("travel-story.theme-ocean", "Ocean Blue")}</SelectItem>
                <SelectItem value="sunset">{t("travel-story.theme-sunset", "Sunset Orange")}</SelectItem>
                <SelectItem value="forest">{t("travel-story.theme-forest", "Forest Green")}</SelectItem>
                <SelectItem value="minimal">{t("travel-story.theme-minimal", "Minimal")}</SelectItem>
              </SelectContent>
            </Select>
          </div>

          {/* Visibility */}
          <div className="flex flex-col gap-2">
            <Label>{t("common.visibility", "Visibility")}</Label>
            <Select value={visibility} onValueChange={(v) => setVisibility(v as typeof visibility)}>
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="PRIVATE">{t("memo.visibility.private", "Private")}</SelectItem>
                <SelectItem value="PROTECTED">{t("memo.visibility.protected", "Protected")}</SelectItem>
                <SelectItem value="PUBLIC">{t("memo.visibility.public", "Public")}</SelectItem>
              </SelectContent>
            </Select>
          </div>

          {/* Auto-generate toggle */}
          <div className="flex items-center gap-2">
            <input
              type="checkbox"
              id="autoGenerate"
              checked={autoGenerate}
              onChange={(e) => setAutoGenerate(e.target.checked)}
              className="rounded border-gray-300"
            />
            <Label htmlFor="autoGenerate" className="text-sm font-normal cursor-pointer">
              {t("travel-story.auto-refine", "Automatically refine content and select best photos")}
            </Label>
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={handleClose}>
            {t("common.cancel", "Cancel")}
          </Button>
          <Button onClick={handleCreate} disabled={!title || selectedMemoNames.length === 0 || isGenerating}>
            {isGenerating ? (
              <>
                <Loader2Icon className="w-4 h-4 mr-2 animate-spin" />
                {t("travel-story.generating", "Generating...")}
              </>
            ) : (
              <>
                <SparklesIcon className="w-4 h-4 mr-2" />
                {t("travel-story.create", "Create Story")}
              </>
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};

export default GenerateTravelStoryDialog;
