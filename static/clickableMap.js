let clickX;
let clickY;

document.getElementById("clickOverlay").addEventListener("click", function (event) {
    const rect = this.getBoundingClientRect();
    clickX = Math.floor(event.clientX - rect.left);
    clickY = Math.floor(event.clientY - rect.top);
    console.log(`Clicked at (${clickX}, ${clickY})`);
});

document.body.addEventListener("htmx:configRequest", (event) => {
    event.detail.parameters.x = clickX;
    event.detail.parameters.y = clickY;
    event.detail.parameters.color = document.getElementById("colorPicker").value;
    console.log(`Sending request for (${clickX}, ${clickY}) with color ${event.detail.parameters.color}`);
});
