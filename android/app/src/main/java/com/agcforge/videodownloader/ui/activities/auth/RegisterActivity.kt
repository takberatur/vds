package com.agcforge.videodownloader.ui.activities.auth

import android.content.Intent
import android.os.Bundle
import android.util.Patterns
import android.view.View
import androidx.appcompat.app.AppCompatActivity
import androidx.lifecycle.lifecycleScope
import com.agcforge.videodownloader.data.api.VideoDownloaderRepository
import com.agcforge.videodownloader.databinding.ActivityRegisterBinding
import com.agcforge.videodownloader.utils.Resource
import com.agcforge.videodownloader.utils.showToast
import kotlinx.coroutines.launch

class RegisterActivity : AppCompatActivity() {

    private lateinit var binding: ActivityRegisterBinding
    private val repository = VideoDownloaderRepository()

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        binding = ActivityRegisterBinding.inflate(layoutInflater)
        setContentView(binding.root)

        setupToolbar()
        setupListeners()
    }

    private fun setupToolbar() {
        binding.toolbar.setNavigationOnClickListener {
            finish()
        }
    }

    private fun setupListeners() {
        binding.apply {
            btnRegister.setOnClickListener {
                val fullName = etFullName.text.toString().trim()
                val email = etEmail.text.toString().trim()
                val password = etPassword.text.toString().trim()
                val confirmPassword = etConfirmPassword.text.toString().trim()

                if (validateInput(fullName, email, password, confirmPassword)) {
                    if (cbTerms.isChecked) {
                        register(fullName, email, password)
                    } else {
                        showToast("Please accept terms and conditions")
                    }
                }
            }

            tvSignIn.setOnClickListener {
                finish()
            }
        }
    }

    private fun validateInput(
        fullName: String,
        email: String,
        password: String,
        confirmPassword: String
    ): Boolean {
        var isValid = true

        // Validate full name
        if (fullName.isEmpty()) {
            binding.tilFullName.error = "Full name is required"
            isValid = false
        } else if (fullName.length < 3) {
            binding.tilFullName.error = "Name must be at least 3 characters"
            isValid = false
        } else {
            binding.tilFullName.error = null
        }

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
        } else if (password.length < 8) {
            binding.tilPassword.error = "Password must be at least 8 characters"
            isValid = false
        } else if (!password.matches(".*[A-Z].*".toRegex())) {
            binding.tilPassword.error = "Password must contain at least one uppercase letter"
            isValid = false
        } else if (!password.matches(".*[0-9].*".toRegex())) {
            binding.tilPassword.error = "Password must contain at least one number"
            isValid = false
        } else {
            binding.tilPassword.error = null
        }

        // Validate confirm password
        if (confirmPassword.isEmpty()) {
            binding.tilConfirmPassword.error = "Please confirm your password"
            isValid = false
        } else if (password != confirmPassword) {
            binding.tilConfirmPassword.error = "Passwords do not match"
            isValid = false
        } else {
            binding.tilConfirmPassword.error = null
        }

        return isValid
    }

    private fun register(fullName: String, email: String, password: String) {
        lifecycleScope.launch {
            showLoading(true)

            try {
                val requestBody = mapOf(
                    "full_name" to fullName,
                    "email" to email,
                    "password" to password
                )

                val result = repository.register(requestBody)

                result.onSuccess { authResponse ->
                    showLoading(false)
                    showToast("Registration successful!")

                    // Navigate to verify email
                    startActivity(Intent(this@RegisterActivity, VerifyEmailActivity::class.java).apply {
                        putExtra("email", email)
                    })
                    finish()
                }.onFailure { error ->
                    showLoading(false)
                    showToast(error.message ?: "Registration failed")
                }

            } catch (e: Exception) {
                showLoading(false)
                showToast("Registration failed: ${e.message}")
            }
        }
    }

    private fun showLoading(isLoading: Boolean) {
        binding.apply {
            progressBar.visibility = if (isLoading) View.VISIBLE else View.GONE
            btnRegister.isEnabled = !isLoading
            etFullName.isEnabled = !isLoading
            etEmail.isEnabled = !isLoading
            etPassword.isEnabled = !isLoading
            etConfirmPassword.isEnabled = !isLoading
        }
    }
}