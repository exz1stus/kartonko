import PerspectiveCard from './PerspectiveCard';
import { noUse } from '@/app/AudioEffects';
import { ImageData, ImageCardServer } from './ImageCardServer';
import React, { useRef } from "react";

interface Props extends ImageData {
    className?: string;
    style?: React.CSSProperties;
}

const ImageCard: React.FC<Props> = ({ filename, tags, className, width, height, style }) => {
    const selfRef = useRef<HTMLDivElement>(null);

    const onMouseEnter = () => {

    }

    const onMouseLeave = () => {

    }

    const onClick = () => {
        noUse();
    }

    const onLoad = () => {
        selfRef.current?.classList.remove("opacity-0");
        selfRef.current?.clientHeight;
    }

    return (
        <div
            ref={selfRef}
            className={className + " opacity-0 "}
            style={style}
            onClick={onClick}
            onMouseEnter={onMouseEnter}
            onMouseLeave={onMouseLeave}
        >
            <PerspectiveCard>
                <ImageCardServer filename={filename} tags={tags} onLoad={onLoad} width={width} height={height} />
            </PerspectiveCard >
        </div>
    )
}

export { type ImageData, ImageCard }
