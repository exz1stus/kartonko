import { useEffect, useState } from "react";
import Image from "next/image";
import { CheckCircle2, Tags } from "lucide-react";
import { toast } from "sonner";
import useUploadStore, { UploadItem } from "@/hooks/useUploadStore";
import TagSelector from "../Tags/TagSelector";
import Scrollbar from "../template/Scrollbar";
import { cn } from "@/lib/utils";
import { beginSelection, extendSelection, SelectionGesture } from "@/lib/image/rangeSelection";

interface Props {
    images: UploadItem[];
    currentIndex: number;
    onIndexChange: (index: number) => void;
    selectedIndices: number[];
    setSelectedIndices: React.Dispatch<React.SetStateAction<number[]>>;
    getFileUrl: (file: File) => string;
}

const ImageCarousel = ({
    images,
    currentIndex,
    onIndexChange,
    selectedIndices,
    setSelectedIndices,
    getFileUrl,
}: Props) => {
    const updateMetadata = useUploadStore((state) => state.updateMetadata);
    const [tags, setTags] = useState<string[]>([]);
    const [newTags, setNewTags] = useState<string[]>([]);

    const [lastSelectedIndex, setLastSelectedIndex] = useState<number | null>(
        null,
    );
    const [selectionGesture, setSelectionGesture] = useState<SelectionGesture<number> | null>(null);

    const applyCommonTags = () => {
        if (selectedIndices.length === 0) return;
        const editableIndices = selectedIndices.filter((index) => images[index] && !images[index].uploadedImageId);
        if (editableIndices.length === 0) {
            toast.warning("Uploaded images cannot receive new upload tags");
            return;
        }

        editableIndices.forEach((index) => {
            const currentItem = images[index];
            if (!currentItem) return;

            updateMetadata(index, {
                tags: [...new Set([...currentItem.tags, ...tags])],
                newTags: [...new Set([...currentItem.newTags, ...newTags])],
            });
        });

        toast.success(`Applied tags to ${editableIndices.length} images`);
        setTags([]);
        setNewTags([]);
        setSelectedIndices([]);
    };

    // --- KEYBOARD SHORTCUTS ---
    useEffect(() => {
        const handleKeyDown = (e: KeyboardEvent) => {
            if (
                e.target instanceof HTMLInputElement ||
                e.target instanceof HTMLTextAreaElement
            ) {
                return;
            }

            if (e.ctrlKey || e.metaKey) {
                if (e.key.toLowerCase() === "a") {
                    e.preventDefault();
                    setSelectedIndices(images.map((_, i) => i));
                } else if (e.key.toLowerCase() === "d") {
                    e.preventDefault();
                    setSelectedIndices([]);
                }
            }
        };

        window.addEventListener("keydown", handleKeyDown);
        return () => window.removeEventListener("keydown", handleKeyDown);
    }, [images, setSelectedIndices]);

    // --- INTEGRATED MOUSE INTERACTION SYSTEM (Windows Explorer Style) ---
    const handleMouseDown = (index: number, e: React.MouseEvent) => {
        if (e.button !== 0) return; // Left click only

        onIndexChange(index);
        const result = beginSelection({
            order: images.map((_, itemIndex) => itemIndex),
            target: index,
            selected: selectedIndices,
            anchor: lastSelectedIndex,
            shift: e.shiftKey,
            additive: e.ctrlKey || e.metaKey,
        });
        setSelectedIndices(result.selected);
        setSelectionGesture(result.gesture);
        setLastSelectedIndex(result.anchor);
    };

    const handleMouseEnter = (index: number) => {
        if (!selectionGesture) return;
        setSelectedIndices(extendSelection(selectionGesture, index));
    };

    const handleCheckboxMouseDown = (index: number, e: React.MouseEvent) => {
        const selected = selectedIndices.includes(index);
        if (selected) {
            setSelectedIndices((prev) => prev.filter((i) => i !== index));
        } else {
            setSelectedIndices((prev) => Array.from(new Set([...prev, index])));
        }
    };

    useEffect(() => {
        const handleMouseUp = () => {
            setSelectionGesture(null);
        };
        window.addEventListener("mouseup", handleMouseUp);
        return () => window.removeEventListener("mouseup", handleMouseUp);
    }, []);

    return (
        <div
            className={cn(
                "flex flex-col gap-4 p-4 border-t w-full min-h-0 select-none shrink-0",
                "lg:border-t-0 lg:border-l lg:w-[20%]",
            )}
        >
            <div className="flex flex-col gap-1">
                <span className="text-gray-400 text-xs">
                    {selectedIndices.length} selected for batch tags
                </span>
            </div>

            {selectedIndices.length > 0 && (
                <div className="flex flex-col gap-2 bg-surface-0 p-2 border rounded-xl animate-in duration-200 fade-in">
                    <div className="flex items-center gap-1.5 font-medium text-xs">
                        <Tags size={14} />
                        <span>Add Common Tags</span>
                    </div>
                    <TagSelector
                        tags={tags}
                        newTags={newTags}
                        removeTag={(tag: string) => {
                            setTags((prevTags) =>
                                prevTags.filter((t) => t !== tag),
                            );
                        }}
                        onTagsUpdate={setTags}
                        removeNewTag={(tag: string) => {
                            setNewTags((prevNewTags) =>
                                prevNewTags.filter((t) => t !== tag),
                            );
                        }}
                        onNewTagsUpdate={setNewTags}
                    />
                    <button
                        onClick={applyCommonTags}
                        className="bg-primary-10 hover:bg-primary-10/90 py-1 rounded-lg w-full font-medium text-black text-xs transition-colors"
                    >
                        Apply to selected
                    </button>
                </div>
            )}

            <Scrollbar className="flex flex-row lg:flex-col flex-1 gap-2 pr-0 lg:pr-1 pb-2 lg:pb-0 min-h-20 lg:min-h-0 overflow-x-auto lg:overflow-x-hidden lg:overflow-y-auto">
                {images.map((img, idx) => {
                    const isSelected = selectedIndices.includes(idx);
                    const isActive = idx === currentIndex;
                    const tagsCount = img.newTags.length + img.tags.length;
                    return (
                        <div
                            key={idx}
                            onMouseDown={(e) => handleMouseDown(idx, e)}
                            onMouseEnter={() => handleMouseEnter(idx)}
                            className={`relative group shrink-0 w-16 h-16 lg:w-full lg:h-20 rounded-xl overflow-hidden border-2 cursor-pointer transition-all ${
                                isActive
                                    ? "border-white"
                                    : isSelected
                                      ? "border-blue-500"
                                      : "border-transparent hover:border-surface-30"
                            }`}
                        >
                            <Image
                                src={getFileUrl(img.file)}
                                alt={img.name}
                                fill
                                sizes="96px"
                                className="object-cover pointer-events-none"
                                unoptimized
                                draggable={false}
                            />

                            {/* Checkbox badge overlay matches base row selection rules */}
                            <div
                                onMouseDown={(e) => {
                                    e.stopPropagation();
                                    handleCheckboxMouseDown(idx, e);
                                }}
                                className={`absolute top-1 left-1 w-5 h-5 rounded-md flex items-center justify-center backdrop-blur-sm transition-all ${
                                    isSelected
                                        ? "bg-blue-500 text-white opacity-100"
                                        : "hidden"
                                }`}
                            >
                                <CheckCircle2
                                    size={14}
                                    className={
                                        isSelected
                                            ? "fill-white text-blue-500"
                                            : ""
                                    }
                                />
                            </div>

                            <div className="right-1 bottom-1 absolute bg-black/60 px-1 rounded text-[10px] pointer-events-none">
                                {idx + 1}
                            </div>

                            <div className="hidden bottom-1 left-0.5 absolute lg:flex bg-black/60 px-1 rounded max-w-[70%] text-[10px] truncate pointer-events-none">
                                {img.name}
                            </div>

                            {tagsCount > 0 && (
                                <div className="hidden top-1 right-1 absolute lg:flex items-center gap-1 bg-black/60 px-1.5 py-0.5 rounded text-[10px] whitespace-nowrap pointer-events-none">
                                    <Tags size={12} />
                                    <span>{tagsCount}</span>
                                </div>
                            )}
                        </div>
                    );
                })}
            </Scrollbar>
        </div>
    );
};

export default ImageCarousel;
