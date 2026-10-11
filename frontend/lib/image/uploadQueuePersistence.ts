import { UploadItem } from "@/hooks/useUploadStore";

const DATABASE_NAME = "kartonko-upload-queue";
const STORE_NAME = "queue";
const QUEUE_KEY = "items";

function openDatabase(): Promise<IDBDatabase> {
    return new Promise((resolve, reject) => {
        const request = indexedDB.open(DATABASE_NAME, 1);
        request.onupgradeneeded = () => {
            request.result.createObjectStore(STORE_NAME);
        };
        request.onsuccess = () => resolve(request.result);
        request.onerror = () => reject(request.error);
    });
}

export async function loadUploadQueue(ownerId: number): Promise<UploadItem[]> {
    const database = await openDatabase();
    return new Promise((resolve, reject) => {
        const transaction = database.transaction(STORE_NAME, "readonly");
        const request = transaction.objectStore(STORE_NAME).get(`${QUEUE_KEY}:${ownerId}`);
        request.onsuccess = () => resolve(Array.isArray(request.result) ? request.result as UploadItem[] : []);
        request.onerror = () => reject(request.error);
        transaction.oncomplete = () => database.close();
        transaction.onerror = () => database.close();
    });
}

export async function saveUploadQueue(ownerId: number, items: UploadItem[]): Promise<void> {
    const database = await openDatabase();
    return new Promise((resolve, reject) => {
        const transaction = database.transaction(STORE_NAME, "readwrite");
        transaction.objectStore(STORE_NAME).put(items, `${QUEUE_KEY}:${ownerId}`);
        transaction.oncomplete = () => {
            database.close();
            resolve();
        };
        transaction.onerror = () => {
            database.close();
            reject(transaction.error);
        };
        transaction.onabort = () => {
            database.close();
            reject(transaction.error);
        };
    });
}
