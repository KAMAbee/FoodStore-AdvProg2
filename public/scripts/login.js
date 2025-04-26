document.addEventListener("DOMContentLoaded", function () {
  const loginForm =
    document.querySelector(".main__form") || document.querySelector("form");
  const loginMessage =
    document.getElementById("login-message") || document.createElement("div");

  if (!document.getElementById("login-message")) {
    loginMessage.id = "login-message";
    loginMessage.style.margin = "10px 0";
    if (loginForm) {
      loginForm.parentNode.insertBefore(loginMessage, loginForm.nextSibling);
    }
  }

  const token = localStorage.getItem("token");
  console.log("Login page loaded, existing token:", token ? "exists" : "none");

  if (token) {
    window.location.href = "/profile";
    return;
  }

  if (loginForm) {
    loginForm.addEventListener("submit", function (e) {
      e.preventDefault();
      console.log("Login form submitted");

      const usernameInput =
        document.getElementById("username") ||
        document.querySelector('input[name="username"]');
      const passwordInput =
        document.getElementById("password") ||
        document.querySelector('input[name="password"]');

      if (!usernameInput || !passwordInput) {
        console.error("Username or password input not found");
        loginMessage.textContent = "Form error: Input fields not found";
        loginMessage.style.color = "red";
        return;
      }

      const username = usernameInput.value;
      const password = passwordInput.value;

      if (!username || !password) {
        loginMessage.textContent = "Fill all fields";
        loginMessage.style.color = "red";
        return;
      }

      fetch("/api/users/login", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Accept: "application/json",
        },
        body: JSON.stringify({ username, password }),
      })
        .then((response) => {
          console.log("Login response status:", response.status);
          if (!response.ok) {
            return response.text().then((text) => {
              try {
                const errorData = JSON.parse(text);
                throw new Error(
                  errorData.error || "Invalid username or password"
                );
              } catch (e) {
                throw new Error("Invalid username or password");
              }
            });
          }
          return response.json();
        })
        .then((data) => {
          console.log("Login response data:", data);

          if (!data.token) {
            throw new Error("No token received from server");
          }

          const token = data.token.trim();
          localStorage.setItem("userId", data.user.id);
          localStorage.setItem("username", data.user.username);
          localStorage.setItem("token", token);

          document.cookie = `auth_token=${token};path=/;max-age=86400`;

          window.location.href = "/profile";
        })
        .catch((error) => {
          console.error("Login error:", error);
          loginMessage.textContent = error.message || "Login failed";
          loginMessage.style.color = "red";
        });
    });
  } else {
    console.error("Login form not found");
  }
});
