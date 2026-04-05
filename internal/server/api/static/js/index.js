
console.log("Hello from static js");

function hideSections() {
    const elements = document.querySelectorAll('.section');
    elements.forEach(element => {
        element.style.display = 'none';
    });
}

function showSection(selector) {
    hideSections();
    document.querySelector(selector).style.display = 'block';

}

async function addFeeds() {
    let feeds = document.getElementById("feed-input").value
    document.getElementById("feed-add-button").disabled = true;
    document.getElementById("feed-input").disabled = true;
    let feedList = feeds.split("\n");

    updateProgress(feedList);

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

    let list = document.getElementById("feed_add_output");
    while (true) {
        const { done, value } = await reader.read();
        if (done) {
            document.getElementById("feed-add-button").disabled = false;
            document.getElementById("feed-input").disabled = false;
            break;
        }
        if ((value === undefined)||(value.length === 0)) {
            continue;
        }

        accumulatedChunks += decoder.decode(value, { stream: true });

        const lines = accumulatedChunks.split('\n');
        for (let i = 0; i < lines.length - 1; i++) {
            updateProgress(feedList);
            try {
                const jsonObject = JSON.parse(lines[i]);
                if (jsonObject.url === "") {
                    console.log("pong!");
                    continue;
                }

                console.log('Parsed object:', jsonObject);
                if (jsonObject.error != null) {
                    list.innerHTML += `<li class="feed_item"><span style="color:red; font-weight:bold;">!</span><a target="_blank" href="${jsonObject.url}">${jsonObject.url}</a></li>`;
                    continue;
                }
                list.innerHTML += `<li  class="feed_item">&#x2705;<a target="_blank" href="${jsonObject.url}">${jsonObject.title}</a></li>`;
            } catch (error) {
                console.error('Error parsing JSON chunk:', error);
                console.log("Chunk:", lines[i]);
            }
        }
        accumulatedChunks = lines[lines.length - 1]; // Keep the last, incomplete line
    }
}

function updateProgress(feedList) {
    console.log("updating progress");
    let feedItemList = document.querySelectorAll(".feed_item");
    let listCount = (feedItemList === null) ? 0 : feedItemList.length;
    document.getElementById("progress").innerHTML = `${listCount} of ${feedList.length}`;
}

async function refreshFeeds() {
    document.getElementById("feed_list_output").innerHTML = "";
    fetch("/api/feeds")
        .then(response => response.json())
        .then(data => {
            if ((data!==null)&&(data.length !== 0)) {
                data.forEach(feed => {
                    document.getElementById("feed_list_output").innerHTML += `<li><a target="_blank" href="${feed.url}">${feed.title}</a></li>`;
                })
            }
        })
        .catch(error => console.error('Error:', error));

}

async function search() {

    document.getElementById("feed_search_output").innerHTML = "";
    fetch("/api/feeds/search",{
        method: "POST",
        body: JSON.stringify({
            "query": document.getElementById("search-input").value
        }),
    }).then(response => response.json())
        .then(data => {
            data.results.forEach(feed => {
                console.log(feed);
                document.getElementById("feed_search_output").innerHTML += `<li><a target="_blank" href="${feed.article.link}">${feed.article.title}</a> - (${feed.feed_title})</li>`;
            })
        }).catch(error => console.error('Error:', error));
    console.log("searching on feeds");
}

async function ask() {

    document.getElementById("question_output_question").innerHTML = document.getElementById("question-input").value;
    document.getElementById("question_output_answer").innerHTML = "Waiting for response...";
    document.getElementById("question-input").disabled = true;
    document.getElementById("question-button").disabled = true;
    fetch("/api/question",{
        method: "POST",
        body: JSON.stringify({
            "question": document.getElementById("question-input").value
        }),
    }).then(response => response.json())
    .then(resp => {
        document.getElementById("question_output_answer").innerHTML = `<span>${resp.answer}</span>`;
        document.getElementById("question-input").disabled = false;
        document.getElementById("question-button").disabled = false;
    }).catch(error => console.error('Error:', error));
    console.log("searching on feeds");
}

document.addEventListener('DOMContentLoaded', (event) => {
    console.log("load function called");
    console.log(JSON.stringify(event));
    document.getElementById("feed-add-button").onclick = addFeeds;

    refreshFeeds().then(() => console.log("refreshed feeds"));
    showSection("#search_feed_container");
});
