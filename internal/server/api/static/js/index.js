
console.log("Hello from static js");

async function streamJsonData(url) {
    let feeds = document.getElementById("feed-input").value
    document.getElementById("feed-add-button").disabled = true;
    document.getElementById("feed-input").disabled = true;
    let feedList = feeds.split("\n")
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

    document.getElementById("progress").textContent = `0 / ${feedList.length}`;
    let list = document.getElementById("feed_add_output");
    let counter = 1;
    while (true) {
        const { done, value } = await reader.read();
        if (done) {
            document.getElementById("feed-add-button").disabled = false;
            document.getElementById("feed-input").disabled = false;
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
            document.getElementById("progress").innerHTML = `${counter} of ${feedList.length}`;
            try {
                const jsonObject = JSON.parse(lines[i]);
                console.log('Parsed object:', jsonObject);
                if (jsonObject.error != null) {
                    list.innerHTML += `<li><span style="color:red; font-weight:bold;">!</span><a target="_blank" href="${jsonObject.url}">${jsonObject.url}</a></li>`;
                    continue;
                }
                list.innerHTML += `<li>&#x2705;<a target="_blank" href="${jsonObject.url}">${jsonObject.title}</a></li>`;
            } catch (error) {
                console.error('Error parsing JSON chunk:', error);
                console.log("Chunk:", lines[i]);
            }
        }
        accumulatedChunks = lines[lines.length - 1]; // Keep the last, incomplete line
        counter++;
    }
}

document.addEventListener('DOMContentLoaded', (event) => {
    console.log("load function called");
    console.log(JSON.stringify(event));
    document.getElementById("feed-add-button").onclick = streamJsonData;
});