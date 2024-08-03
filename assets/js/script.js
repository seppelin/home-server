function twoD(num) {
    return num.toString().padStart(2, '0');
}

function getCountDown(time) {
    var nums = time.split(":").map(Number)
    var now = new Date().getTime();
    now += nums[0] * 60 * 60 * 1000 + nums[1] * 60 * 1000 + nums[2] * 1000 + 500;
    return now;
}

function startManualTime() {
    console.log("Start Manual Time Update");
    var el = document.getElementById("w-manual-time");
    var countDown = getCountDown(el.innerHTML);
    setTimeout(updateManualTime, 1000, el, countDown);
}

function updateManualTime(el, countDown) {
    var now = new Date().getTime();
    var distance = countDown - now;
    if (distance < 0) {
        console.log("End Time Update: Too small")
        return;
    }

    var hours = Math.floor((distance % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60));
    var minutes = Math.floor((distance % (1000 * 60 * 60)) / (1000 * 60));
    var seconds = Math.floor((distance % (1000 * 60)) / 1000);
    if (!el || !document.body.contains(el)) {
        console.log("End Time Update: No Element")
        return
    }
    el.innerHTML = twoD(hours) + ":" + twoD(minutes) + ":" + twoD(seconds);
    setTimeout(updateManualTime, 1000, el, countDown);
}
