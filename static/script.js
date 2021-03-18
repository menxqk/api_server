const login = document.getElementById('login');
const loginModal = document.getElementById('login-modal');
const close = document.querySelector('.close');
const btnSubmit = document.querySelector('.btn-submit');
const loaderModal = document.getElementById('loader-modal');
const username = document.getElementById('username');
const password = document.getElementById('password');

login.addEventListener('click', (e) => {
    e.preventDefault()
    loginModal.classList.add('show');
    username.focus();
});

close.addEventListener('click', () => {
    loginModal.classList.remove('show');
});

btnSubmit.addEventListener('click', () => {
    doLogin();
});

loginModal.addEventListener('keypress', (e) => {
    if (e.key === 'Enter' && e.target === password) {
        doLogin();
    } else if (e.key === 'Enter' && e.target === username) {
        password.focus();
    }
})

function doLogin() {
    loaderModal.classList.add('show');
    loginModal.classList.remove('show');
    setTimeout(() => {
        loaderModal.classList.remove('show');
    }, 15000);
    const user = username.value;
    const pass = password.value;
    authenticate(user, pass);
}

async function authenticate(user, pass) {
    const config = {
        body: "username=" + user +"&password=" + pass,
        credentials: "include",
        headers: {
            "Content-Type": "application/x-www-form-urlencoded",
        },
        method: "post"
    };
    const res = await fetch('/login', config);
    const data = await res.json();
    
    console.log('data', data);
    if (data.result === "ok") {
        window.location.replace(data.goto);
    } else if (data.result === "failed") {
        loaderModal.classList.remove('show');
        loginModal.classList.add('show');
        alert(data.result, data.message);
    } else {

    }
}
