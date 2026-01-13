import { CheckCircle2Icon, CircleIcon } from "lucide-react";
import { useMemoSelection } from "@/contexts/MemoSelectionContext";
import { cn } from "@/lib/utils";
import type { Memo } from "@/types/proto/api/v1/memo_service_pb";

interface Props {
  memo: Memo;
  className?: string;
}

const MemoSelectionCheckbox = ({ memo, className }: Props) => {
  const { isSelectionMode, isMemoSelected, toggleMemoSelection } = useMemoSelection();

  if (!isSelectionMode) {
    return null;
  }

  const isSelected = isMemoSelected(memo.name);

  return (
    <button
      type="button"
      onClick={(e) => {
        e.stopPropagation();
        e.preventDefault();
        toggleMemoSelection(memo);
      }}
      className={cn(
        "flex items-center justify-center w-6 h-6 rounded-full transition-colors",
        isSelected ? "text-primary" : "text-muted-foreground hover:text-foreground",
        className,
      )}
    >
      {isSelected ? <CheckCircle2Icon className="w-5 h-5" /> : <CircleIcon className="w-5 h-5" />}
    </button>
  );
};

export default MemoSelectionCheckbox;
