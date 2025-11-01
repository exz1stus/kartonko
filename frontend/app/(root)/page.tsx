import GalleryServer from "@/components/Gallery/GalleryServer";

export default function Home() {
    return (
        <GalleryServer initialFetchSize={70} />
    );
}
