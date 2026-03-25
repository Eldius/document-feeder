
console.log("Hello from static js");

document.addEventListener('DOMContentLoaded', (event) => {
    console.log("load function called");
    console.log(JSON.stringify(event));
    document.getElementById("feed-add-button").onclick = function(evt) {
        alert("clicked");
    }
});