import { createContext, useCallback, useContext, useMemo, useState, type ReactNode } from "react";
import type { Memo } from "@/types/proto/api/v1/memo_service_pb";

interface MemoSelectionContextValue {
  selectedMemos: Map<string, Memo>;
  isSelectionMode: boolean;
  toggleSelectionMode: () => void;
  selectMemo: (memo: Memo) => void;
  deselectMemo: (memoName: string) => void;
  toggleMemoSelection: (memo: Memo) => void;
  isMemoSelected: (memoName: string) => boolean;
  clearSelection: () => void;
  selectAll: (memos: Memo[]) => void;
  selectedCount: number;
}

const MemoSelectionContext = createContext<MemoSelectionContextValue | null>(null);

export function MemoSelectionProvider({ children }: { children: ReactNode }) {
  const [selectedMemos, setSelectedMemos] = useState<Map<string, Memo>>(new Map());
  const [isSelectionMode, setIsSelectionMode] = useState(false);

  const toggleSelectionMode = useCallback(() => {
    setIsSelectionMode((prev) => {
      if (prev) {
        // Exiting selection mode, clear selection
        setSelectedMemos(new Map());
      }
      return !prev;
    });
  }, []);

  const selectMemo = useCallback((memo: Memo) => {
    setSelectedMemos((prev) => {
      const next = new Map(prev);
      next.set(memo.name, memo);
      return next;
    });
  }, []);

  const deselectMemo = useCallback((memoName: string) => {
    setSelectedMemos((prev) => {
      const next = new Map(prev);
      next.delete(memoName);
      return next;
    });
  }, []);

  const toggleMemoSelection = useCallback((memo: Memo) => {
    setSelectedMemos((prev) => {
      const next = new Map(prev);
      if (next.has(memo.name)) {
        next.delete(memo.name);
      } else {
        next.set(memo.name, memo);
      }
      return next;
    });
  }, []);

  const isMemoSelected = useCallback(
    (memoName: string) => {
      return selectedMemos.has(memoName);
    },
    [selectedMemos],
  );

  const clearSelection = useCallback(() => {
    setSelectedMemos(new Map());
  }, []);

  const selectAll = useCallback((memos: Memo[]) => {
    setSelectedMemos((prev) => {
      const next = new Map(prev);
      for (const memo of memos) {
        next.set(memo.name, memo);
      }
      return next;
    });
  }, []);

  const value = useMemo(
    () => ({
      selectedMemos,
      isSelectionMode,
      toggleSelectionMode,
      selectMemo,
      deselectMemo,
      toggleMemoSelection,
      isMemoSelected,
      clearSelection,
      selectAll,
      selectedCount: selectedMemos.size,
    }),
    [
      selectedMemos,
      isSelectionMode,
      toggleSelectionMode,
      selectMemo,
      deselectMemo,
      toggleMemoSelection,
      isMemoSelected,
      clearSelection,
      selectAll,
    ],
  );

  return <MemoSelectionContext.Provider value={value}>{children}</MemoSelectionContext.Provider>;
}

export function useMemoSelection() {
  const context = useContext(MemoSelectionContext);
  if (!context) {
    throw new Error("useMemoSelection must be used within a MemoSelectionProvider");
  }
  return context;
}
