function nTS(num) {
    if (num < 10) {
        return "0" + num
    }
    return num
}

function updateManualTime() {
    var el = document.getElementById('w-manual-time')
    if (el == null) {
        console.log("End Time Update")
        return
    }
    var nums = el.innerHTML.split(":").map((s) => Number(s))
    var carrier = true
    for (let i = 2; i >= 0; i--) {
        if (carrier) {
            nums[i] -= 1
            if (nums[i] < 0) {
                nums[i] = 60 + nums[i]
                carrier = true
            } else {
                carrier = false
            }
        }
    }
    el.innerHTML = nums.map((s) => nTS(s)).join(":")
    setTimeout(updateManualTime, 1000)
}