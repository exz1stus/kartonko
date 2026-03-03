import GalleryServer from "@/components/Gallery/GalleryServer";

export default function Home() {
    return (
        <>
            <meta
                name="description"
                content="Те, кому не нравятся слова ХУЙ и ПИЗДА, могут идти нахуй. Остальные пруцца!"
            />
            <GalleryServer initialFetchSize={60} />
        </>
    );
}
