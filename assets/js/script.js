// function getCountDown(time) {
//     var nums = time.split(":").map(Number)
//     var now = new Date().getTime();
//     now += nums[0] * 60 * 60 * 1000 + nums[1] * 60 * 1000 + nums[2] * 1000 + 500;
//     return now;
// }

function startCountDown(id) {
    console.log("Start Count Down: " + id);
    var el = document.getElementById(id);
    var countDown = Number(el.innerHTML);
    countDown += new Date().getTime();
    updateCountDown(el, countDown);
}

function updateCountDown(el, countDown) {
    var now = new Date().getTime();
    var distance = countDown - now;
    if (!el || !document.body.contains(el)) {
        console.log("End Time Update: No Element: " + el.id)
        return
    }
    if (distance < 0) {
        console.log("End Time Update: Too small: " + el.id)
        return;
    }

    var hours = Math.floor((distance % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60));
    var minutes = Math.floor((distance % (1000 * 60 * 60)) / (1000 * 60));
    var seconds = Math.floor((distance % (1000 * 60)) / 1000);
    
    el.innerHTML = hours + "h " + minutes + "m " + seconds + "s";
    setTimeout(updateCountDown, distance%1000, el, countDown);
}
