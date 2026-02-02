package com.agcforge.videodownloader.ui.activities.auth

import android.content.Intent
import android.os.Bundle
import android.util.Patterns
import android.view.View
import androidx.activity.result.contract.ActivityResultContracts
import androidx.activity.viewModels
import androidx.appcompat.app.AppCompatActivity
import androidx.lifecycle.lifecycleScope
import com.agcforge.videodownloader.databinding.ActivityLoginBinding
import com.agcforge.videodownloader.service.WebSocketService
import com.agcforge.videodownloader.ui.activities.MainActivity
import com.agcforge.videodownloader.ui.viewmodel.AuthViewModel
import com.agcforge.videodownloader.utils.PreferenceManager
import com.agcforge.videodownloader.utils.Resource
import com.agcforge.videodownloader.utils.showToast
import com.google.android.gms.auth.api.signin.GoogleSignIn
import com.google.android.gms.auth.api.signin.GoogleSignInClient
import com.google.android.gms.auth.api.signin.GoogleSignInOptions
import com.google.android.gms.common.api.ApiException
import kotlinx.coroutines.launch

class LoginActivity : AppCompatActivity() {

    private lateinit var binding: ActivityLoginBinding
    private val viewModel: AuthViewModel by viewModels()
    private lateinit var preferenceManager: PreferenceManager
    private lateinit var googleSignInClient: GoogleSignInClient

	private val googleSignInLauncher = registerForActivityResult(
		ActivityResultContracts.StartActivityForResult()
	) { result ->
		val data = result.data
		val task = GoogleSignIn.getSignedInAccountFromIntent(data)
		try {
			val account = task.getResult(ApiException::class.java)
			val credential = account.idToken
			if (!credential.isNullOrEmpty()) {
				viewModel.loginGoogle(credential)
			} else {
				showToast("Google login failed")
			}
		} catch (e: Exception) {
			showToast(e.message ?: "Google login failed")
		}
	}

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        binding = ActivityLoginBinding.inflate(layoutInflater)
        setContentView(binding.root)

        preferenceManager = PreferenceManager(this)

		val gso = GoogleSignInOptions.Builder(GoogleSignInOptions.DEFAULT_SIGN_IN)
			.requestIdToken(getString(com.agcforge.videodownloader.R.string.google_web_client_id))
			.requestEmail()
			.build()
		googleSignInClient = GoogleSignIn.getClient(this, gso)

        // Check if already logged in
        checkIfLoggedIn()

        setupListeners()
        observeViewModel()
    }

    private fun checkIfLoggedIn() {
        lifecycleScope.launch {
            preferenceManager.authToken.collect { token ->
                if (!token.isNullOrEmpty()) {
                    navigateToMain()
                }
            }
        }
    }

    private fun setupListeners() {
        binding.apply {
            btnLogin.setOnClickListener {
                val email = etEmail.text.toString().trim()
                val password = etPassword.text.toString().trim()

                if (validateInput(email, password)) {
                    viewModel.login(email, password)
                }
            }

            tvSignUp.setOnClickListener {
                startActivity(Intent(this@LoginActivity, RegisterActivity::class.java))
            }

            tvForgotPassword.setOnClickListener {
                startActivity(Intent(this@LoginActivity, ForgotPasswordActivity::class.java))
            }

            btnGoogleLogin.setOnClickListener {
				googleSignInLauncher.launch(googleSignInClient.signInIntent)
            }
        }
    }

    private fun validateInput(email: String, password: String): Boolean {
        var isValid = true

        // Validate email
        if (email.isEmpty()) {
            binding.tilEmail.error = "Email is required"
            isValid = false
        } else if (!Patterns.EMAIL_ADDRESS.matcher(email).matches()) {
            binding.tilEmail.error = "Invalid email format"
            isValid = false
        } else {
            binding.tilEmail.error = null
        }

        // Validate password
        if (password.isEmpty()) {
            binding.tilPassword.error = "Password is required"
            isValid = false
        } else if (password.length < 6) {
            binding.tilPassword.error = "Password must be at least 6 characters"
            isValid = false
        } else {
            binding.tilPassword.error = null
        }

        return isValid
    }

    private fun observeViewModel() {
        lifecycleScope.launch {
            viewModel.loginResult.collect { resource ->
                when (resource) {
                    is Resource.Loading -> {
                        showLoading(true)
                    }
                    is Resource.Success -> {
                        showLoading(false)
                        resource.data?.let { authResponse ->
                            showToast("Login successful!")

                            // Start WebSocket service
                            WebSocketService.start(
                                this@LoginActivity,
                                authResponse.user.id,
                                authResponse.token
                            )

                            navigateToMain()
                        }
                    }
                    is Resource.Error -> {
                        showLoading(false)
                        showToast(resource.message ?: "Login failed")
                    }
                }
            }
        }
    }

    private fun showLoading(isLoading: Boolean) {
        binding.apply {
            progressBar.visibility = if (isLoading) View.VISIBLE else View.GONE
            btnLogin.isEnabled = !isLoading
            btnGoogleLogin.isEnabled = !isLoading
        }
    }

    private fun navigateToMain() {
        startActivity(Intent(this, MainActivity::class.java).apply {
            flags = Intent.FLAG_ACTIVITY_NEW_TASK or Intent.FLAG_ACTIVITY_CLEAR_TASK
        })
        finish()
    }
}
