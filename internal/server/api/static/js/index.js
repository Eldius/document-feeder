
console.log("Hello from static js");

async function streamJsonData(url) {
    document.getElementById("feed-add-button").disabled = true;
    let feeds = document.getElementById("feed-input").value
    let feedList = feeds.replace("\n", ",").split(",")
    const response = await fetch("/api/feeds",
        {
            method: "POST",
            body: JSON.stringify({
                "feeds": feedList
            }),
        })
    if (!response.body) return;

    const reader = response.body.getReader();
    const decoder = new TextDecoder();
    let accumulatedChunks = '';

    while (true) {
        const { done, value } = await reader.read();
        if (done) {
            // Process any remaining data if needed
            break;
        }

        // Decode the Uint8Array chunk to a string
        accumulatedChunks += decoder.decode(value, { stream: true });

        // You now need a way to parse valid JSON objects from 'accumulatedChunks'
        // as they become available. This is where a specialized library
        // or a specific protocol (like NDJSON) is necessary.
        // For example, if the server sends newline-delimited JSON (NDJSON):
        const lines = accumulatedChunks.split('\n');
        for (let i = 0; i < lines.length - 1; i++) {
            try {
                const jsonObject = JSON.parse(lines[i]);
                console.log('Parsed object:', jsonObject);
                // Handle the parsed object (e.g., update UI)
            } catch (error) {
                console.error('Error parsing JSON chunk:', error);
            }
        }
        accumulatedChunks = lines[lines.length - 1]; // Keep the last, incomplete line
    }
}

document.addEventListener('DOMContentLoaded', (event) => {
    console.log("load function called");
    console.log(JSON.stringify(event));
    document.getElementById("feed-add-button").onclick = streamJsonData;
});