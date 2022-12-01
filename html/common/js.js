{{ define "common-js" }}

function updateSessionTimer() {
    const timerId = "session-timer-value"
    const diffMillis = getSessionDeadline() - Date.now(); // getSessionDeadline should be provided in target template
    const diffSeconds = Math.floor(diffMillis / 1000);
    const minutesLeft = Math.floor(diffSeconds / 60);
    const secondsLeft = diffSeconds - 60 * minutesLeft;

    timer = ''
    if (minutesLeft < 10) {
        timer += '0'
    }
    timer += minutesLeft + ':'

    if (secondsLeft < 10) {
        timer += '0'
    }
    timer += secondsLeft

    document.getElementById(timerId).innerHTML = timer
}

var t = setInterval(updateSessionTimer, 1000);

{{ end }}
