import PerspectiveCard from "./PerspectiveCard";
import ImageCardServer from "./ImageCardServer";
import React, { useRef } from "react";
import { useRouter } from "next/navigation";
import ImageMetadata from "@/app/lib/image";

interface Props extends ImageMetadata {
    className?: string;
    style?: React.CSSProperties;
}

const ImageCard: React.FC<Props> = ({
    filename,
    tags,
    format,
    className,
    width,
    height,
    style,
}) => {
    const selfRef = useRef<HTMLDivElement>(null);
    const router = useRouter();
    const onClick = () => {
        router.push(`/image/${filename}`);
    };

    const onLoad = () => {
        selfRef.current?.classList.remove("opacity-0");
        selfRef.current?.clientHeight;
    };

    return (
        <div ref={selfRef} className={className + " opacity-0 "} style={style} onClick={onClick}>
            <PerspectiveCard>
                <ImageCardServer
                    filename={filename}
                    format={format}
                    tags={tags}
                    onLoad={onLoad}
                    width={width}
                    height={height}
                />
            </PerspectiveCard>
        </div>
    );
};

export default ImageCard;
