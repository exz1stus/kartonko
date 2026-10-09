"use client";
import { ImageMetadata } from "@/lib/api/generated/model/imageMetadata";
import { Pencil, Save, Trash2, X } from "lucide-react";
import { useCallback, useEffect, useState } from "react";
import { toast } from "sonner";
import { useRouter } from "next/navigation";
import {
    deleteImage as deleteImageRequest,
    getImageByName,
    postTagsBatch,
    patchImage,
} from "@/lib/api/generated/client";
import TagSelector from "@/components/Tags/TagSelector";
import { sanitizeName } from "@/lib/sanitizeName";
import ApiError from "@/lib/api/error";

interface Props {
    image: ImageMetadata;
    hasPermission: boolean;
    onDelete?: () => void;
}

const EditImage = ({ image, hasPermission, onDelete }: Props) => {
    const router = useRouter();
    const [editing, setEditing] = useState(false);
    const [filename, setFilename] = useState(image.filename);
    const [tags, setTags] = useState(image.tags ?? []);
    const [newTags, setNewTags] = useState<string[]>([]);
    const [pending, setPending] = useState(false);

    useEffect(() => {
        if (editing) return;
        setFilename(image.filename);
        setTags(image.tags ?? []);
        setNewTags([]);
    }, [editing, image.filename, image.tags]);

    const save = useCallback(async () => {
        if (pending) return;
        setPending(true);
        try {
            const nextFilename = sanitizeName(filename.trim());
            if (!nextFilename) throw new Error("Image name cannot be empty");

            if (nextFilename !== image.filename) {
                try {
                    const existing = await getImageByName(nextFilename);
                    if (existing.id !== image.id) {
                        throw new Error(`Image name already exists: ${nextFilename}`);
                    }
                } catch (error) {
                    if (!(error instanceof ApiError && error.status === 404)) {
                        throw error;
                    }
                }
            }

            if (newTags.length > 0) {
                const result = await postTagsBatch(
                    { names: newTags },
                    { credentials: "include" },
                );
                if (result.failures?.length) {
                    throw new Error(
                        result.failures
                            .map((failure) => `${failure.name}: ${failure.error}`)
                            .join(", "),
                    );
                }
            }

            const updated = await patchImage(
                image.id,
                { filename: nextFilename, tags: tags.concat(newTags) },
                { credentials: "include" },
            );
            setEditing(false);
            setNewTags([]);
            router.replace(`/image/${encodeURIComponent(updated.filename)}`);
            router.refresh();
        } finally {
            setPending(false);
        }
    }, [filename, image.id, newTags, pending, router, tags]);
    const fetchDelete = useCallback(async () => {
        if (pending) return;
        setPending(true);
        try {
            await deleteImageRequest(image.id, { credentials: "include" });
        } finally {
            setPending(false);
        }
    }, [image.id, pending]);

    const removeImage = useCallback(async () => {
        if (pending) return;
        if (!window.confirm(`Delete image "${image.filename}"?`)) return;
        toast.promise(fetchDelete, {
            loading: "Loading...",
            success: () => {
                onDelete?.();
                router.push("/");
                return `image has been deleted`;
            },
            error: (error) => error.message,
        });
    }, [fetchDelete, image.filename, onDelete, pending, router]);

    if (!hasPermission) return null;

    return (
        <div className="flex flex-col gap-2">
            <div className="flex flex-row items-center gap-2">
                <span className="text-2xl">Edit:</span>
                {!editing && (
                    <button
                        type="button"
                        disabled={pending}
                        onClick={() => setEditing(true)}
                        className="m-2 hover:text-primary disabled:opacity-50"
                        aria-label="Edit image"
                    >
                        <Pencil />
                    </button>
                )}
                <button
                    type="button"
                    disabled={pending}
                    onClick={() => void removeImage()}
                    className="m-2 hover:text-red-500 disabled:opacity-50"
                    aria-label="Delete image"
                >
                    <Trash2 />
                </button>
            </div>
            {editing && (
                <div className="flex flex-col gap-2 min-w-64">
                    <input
                        className="border rounded-md px-2 py-1 bg-transparent"
                        value={filename}
                        onChange={(event) => setFilename(sanitizeName(event.target.value))}
                        aria-label="Image name"
                    />
                    <TagSelector
                        inputStyle="flex flex-wrap gap-1 border rounded-t-md w-full h-full bg-neutral-900 px-3 py-2"
                        tags={tags}
                        removeTag={(tag) => setTags((current) => current.filter((item) => item !== tag))}
                        onTagsUpdate={setTags}
                        newTags={newTags}
                        removeNewTag={(tag) => setNewTags((current) => current.filter((item) => item !== tag))}
                        onNewTagsUpdate={setNewTags}
                    />
                    <div className="flex gap-2">
                        <button
                            type="button"
                            disabled={pending}
                            onClick={() => void toast.promise(save(), {
                                loading: "Saving...",
                                success: "Image updated",
                                error: (error) => error.message,
                            })}
                            className="inline-flex items-center gap-1 border rounded-md px-2 py-1 hover:border-primary"
                        >
                            <Save size={16} /> Save
                        </button>
                        <button
                            type="button"
                            disabled={pending}
                            onClick={() => {
                                setFilename(image.filename);
                                setTags(image.tags ?? []);
                                setNewTags([]);
                                setEditing(false);
                            }}
                            className="inline-flex items-center gap-1 border rounded-md px-2 py-1"
                        >
                            <X size={16} /> Cancel
                        </button>
                    </div>
                </div>
            )}
        </div>
    );
};

export default EditImage;
