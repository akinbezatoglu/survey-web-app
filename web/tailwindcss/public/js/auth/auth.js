function Registration() {
    var name = document.getElementById("name").value;
    var lastname = document.getElementById("lastname").value;
    var email = document.getElementById("email").value;
    var password = document.getElementById("password").value;
    var confirmPassword = document.getElementById("confirmpassword").value;
    var err = document.getElementById("serverMessageBox")
    if (name == "" || lastname == "" || email == "" || password == "" || confirmPassword == "") {
        alert("Tüm bilgileri eksiksiz doldurunuz.")
        return false;
    } 
    if (password != confirmPassword) {
        alert("Parola onaylanmadı.")
        return false;
    }
    if (!ValidateEmail(email)) {
        alert("Girmiş olduğunuz email doğru formatta değil.")
        return false;
    }
    if (!ValidatePassword(password)) {
        alert("Girmiş olduğunuz parola doğru formatta değil.")
        return false;
    }
    data = {
        "name": name,
        "lastname": lastname,
        "email": email,
        "password": password
    }
    console.log(data);
    const response = fetch('http://localhost:8080/api/v1/auth/signup', {
        method: 'POST',
        body: JSON.stringify(data)
    })
    .then(response => response.json())
    .then(data => {
        if (data["_id"] != "") {
            window.localStorage.setItem("token", data["token"]);
            window.location.replace("http://localhost:8000/index.html");
        } else {
            alert("Bu emaile ait kullanıcı bulunmaktadır.")
            return false;
        }
    })
    .catch(error => {
        alert("Beklenmedik bir hata oluştu.")
        return false;
    });
    return false;
}

function Login() {
    var email = document.getElementById("email").value;
    var password = document.getElementById("password").value;
    var err = document.getElementById("serverMessageBox")
    if (email == "" || password == "") {
        alert("Tüm bilgileri eksiksiz doldurunuz.")
        return false;
    } 
    if (!ValidateEmail(email)) {
        alert("Girmiş olduğunuz email doğru formatta değil.")
        return false;
    }
    if (!ValidatePassword(password)) {
        alert("Girmiş olduğunuz parola doğru formatta değil.")
        return false;
    }
    data = {
        "email": email,
        "password": password
    }
    const response = fetch('http://localhost:8080/api/v1/auth/login', {
        method: 'POST',
        body: JSON.stringify(data)
    })
    .then(response => response.json())
    .then(data => {
        if (data != null) {
            window.localStorage.setItem("token", data["token"]);
            window.location.replace("http://localhost:8000/index.html");
        } else {
            alert("Kullanıcı bilgilerinizi hatalı girdiniz.")
            return false;
        }
    })
    .catch(error => {
        alert("Beklenmedik bir hata oluştu.")
        return false;
    });
    return false;
}

function ValidateEmail(input) {
    var validRegex = /^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9-]+(?:\.[a-zA-Z0-9-]+)*$/;
    if (input.match(validRegex)) {
        return true;
    } else {
        return false;
    }
}

function ValidatePassword(input) {
    var passRegex = new RegExp ("^(((?=.*[a-z])(?=.*[A-Z]))|((?=.*[a-z])(?=.*[0-9]))|((?=.*[A-Z])(?=.*[0-9])))(?=.{6,})");
    if (input.match(passRegex)) {
        return true;
    } else {
        return false;
    }
}

window.onload = async function() {
    const response = await fetch('http://localhost:8080/api/v1/auth', {
        method: 'GET',
        headers: {
            'Authorization': 'Bearer ' + window.localStorage.getItem('token')
        }
    })
    .then(response => response.json())
    .then(data => {
        if (data["_id"] != "") {
            window.location.replace("http://localhost:8000/index.html");
        }
    })
    .catch(error => {
        console.log(error);
    });
}