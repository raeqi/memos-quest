import { useState } from "react";
import { BookOpenIcon, CheckSquareIcon, SquareIcon, XIcon } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useMemoSelection } from "@/contexts/MemoSelectionContext";
import GenerateTravelStoryDialog from "./GenerateTravelStoryDialog";
import { useTranslate } from "@/utils/i18n";

const MemoSelectionToolbar = () => {
  const t = useTranslate();
  const { isSelectionMode, toggleSelectionMode, selectedCount, clearSelection } = useMemoSelection();
  const [showGenerateDialog, setShowGenerateDialog] = useState(false);

  if (!isSelectionMode) {
    return (
      <Button variant="outline" size="sm" onClick={toggleSelectionMode} className="gap-2">
        <CheckSquareIcon className="w-4 h-4" />
        {t("travel-story.select-memos", "Select Memos")}
      </Button>
    );
  }

  return (
    <>
      <div className="flex items-center gap-2 px-3 py-2 bg-primary/10 border border-primary/20 rounded-lg">
        <div className="flex items-center gap-2 text-sm">
          {selectedCount > 0 ? (
            <CheckSquareIcon className="w-4 h-4 text-primary" />
          ) : (
            <SquareIcon className="w-4 h-4 text-muted-foreground" />
          )}
          <span>
            <strong>{selectedCount}</strong> {t("travel-story.selected", "selected")}
          </span>
        </div>

        <div className="flex items-center gap-1 ml-auto">
          {selectedCount > 0 && (
            <>
              <Button variant="ghost" size="sm" onClick={clearSelection}>
                {t("common.clear", "Clear")}
              </Button>
              <Button size="sm" onClick={() => setShowGenerateDialog(true)} className="gap-2">
                <BookOpenIcon className="w-4 h-4" />
                {t("travel-story.generate-story", "Generate Story")}
              </Button>
            </>
          )}
          <Button variant="ghost" size="icon" onClick={toggleSelectionMode} className="ml-1">
            <XIcon className="w-4 h-4" />
          </Button>
        </div>
      </div>

      <GenerateTravelStoryDialog open={showGenerateDialog} onOpenChange={setShowGenerateDialog} />
    </>
  );
};

export default MemoSelectionToolbar;
