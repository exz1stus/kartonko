import ImageMetadata from "@/lib/image";
import Gallery from "./Gallery";
import { SearchQuery } from "./ImageSearch";
import { serverFetch } from "@/lib/serverFetch";

const fetchImages = async (intialFetchSize: number, initialQuery: SearchQuery) => {
    const nameQueryString =
        initialQuery.nameContains.length === 0 ? "" : `name=${initialQuery.nameContains}&`;
    const tagsQueryString =
        initialQuery.withTags.length === 0 ? "" : `tags=${JSON.stringify(initialQuery.withTags)}&`;
    const response = await serverFetch(
        `/images?${nameQueryString}${tagsQueryString}cursor=${0}&limit=${intialFetchSize}`,
        {
            cache: "no-store",
        },
    );

    if (!response.ok) throw new Error("Failed to fetch images");
    console.log(`server intial fetching cursor ${0} limit ${intialFetchSize}`);
    const data: ImageMetadata[] = await response.json();
    return data;
};

interface Props {
    initialFetchSize: number;
}

const GalleryServer = async ({ initialFetchSize }: Props) => {
    const INITIAL_QUERY: SearchQuery = {
        nameContains: "",
        withTags: [],
    };
    let images: ImageMetadata[] = [];
    let initReachedEnd = false;
    if (initialFetchSize > 0) {
        try {
            images = await fetchImages(initialFetchSize, INITIAL_QUERY);
            initReachedEnd = images.length < initialFetchSize;
        } catch (error) {
            console.error("Error fetching images:", error);
            images = [];
        }
    }

    return (
        <Gallery
            initialImages={images}
            initReachedEnd={initReachedEnd}
            initialQuery={INITIAL_QUERY}
        />
    );
};

export default GalleryServer;
