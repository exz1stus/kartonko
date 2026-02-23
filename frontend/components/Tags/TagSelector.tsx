import React from "react";
import TagSpan from "./TagSpan";
import TagSelectorField from "./TagSelectorField";

interface Props {
    tags: string[];
    removeTag: (tag: string) => void;
    setTags: (tags: string[]) => void;
}

const TagSelector = ({ tags, removeTag, setTags }: Props) => {
    const [tagFieldActive, setTagFieldActive] = React.useState<boolean>(false);

    return (
        <div className="flex flex-wrap gap-1">
            <TagSpan
                tags={tags}
                removeTag={(t) => removeTag(t)}
                tagStyle={"bg-surface-0 rounded-3xl px-2"}
            />
            <TagSelectorField
                active={tagFieldActive}
                tags={tags}
                onTagsUpdate={(t) => setTags(t)}
            />
        </div>
    );
};

export default TagSelector;
