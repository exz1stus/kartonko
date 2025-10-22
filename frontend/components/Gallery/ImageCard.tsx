import PerspectiveCard from './PerspectiveCard';
import { noUse } from '@/app/AudioEffects';
import { ImageData, ImageCardServer } from './ImageCardServer';

const ImageCard: React.FC<ImageData> = ({ filename, tags }) => {
    const onMouseEnter = () => {
    }

    const onMouseLeave = () => {

    }

    const onClick = () => {
        noUse();
    }

    return (
        <div
            className="hover:z-1"
            onClick={onClick}
            onMouseEnter={onMouseEnter}
            onMouseLeave={onMouseLeave}
        >
            <PerspectiveCard>
                <ImageCardServer filename={filename} tags={tags} />
            </PerspectiveCard >
        </div>
    )
}

export { type ImageData, ImageCard }
