package com.agcforge.videodownloader.ui.activities.auth

import android.content.Intent
import android.net.Uri
import android.os.Bundle
import android.os.CountDownTimer
import android.view.View
import androidx.appcompat.app.AppCompatActivity
import androidx.lifecycle.lifecycleScope
import com.agcforge.videodownloader.data.api.VideoDownloaderRepository
import com.agcforge.videodownloader.databinding.ActivityVerifyEmailBinding
import com.agcforge.videodownloader.utils.showToast
import kotlinx.coroutines.launch

class VerifyEmailActivity : AppCompatActivity() {

    private lateinit var binding: ActivityVerifyEmailBinding
    private val repository = VideoDownloaderRepository()

    private var email: String = ""
    private var resendTimer: CountDownTimer? = null
    private var canResend = true

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        binding = ActivityVerifyEmailBinding.inflate(layoutInflater)
        setContentView(binding.root)

        email = intent.getStringExtra("email") ?: ""

        if (email.isEmpty()) {
            showToast("Email not found")
            finish()
            return
        }

        setupUI()
        setupListeners()
    }

    private fun setupUI() {
        binding.tvEmail.text = email
    }

    private fun setupListeners() {
        binding.apply {
            btnResendEmail.setOnClickListener {
                if (canResend) {
                    resendVerificationEmail()
                } else {
                    showToast("Please wait before resending")
                }
            }

            btnOpenEmailApp.setOnClickListener {
                openEmailApp()
            }

            tvBackToLogin.setOnClickListener {
                navigateToLogin()
            }
        }
    }

    private fun resendVerificationEmail() {
        lifecycleScope.launch {
            showLoading(true)

            try {
                val requestBody = mapOf("email" to email)
                val result = repository.resendVerificationEmail(requestBody)

                result.onSuccess {
                    showLoading(false)
                    showToast("Verification email sent!")
                    startResendTimer()
                }.onFailure { error ->
                    showLoading(false)
                    showToast(error.message ?: "Failed to resend email")
                }

            } catch (e: Exception) {
                showLoading(false)
                showToast("Error: ${e.message}")
            }
        }
    }

    private fun startResendTimer() {
        canResend = false
        binding.btnResendEmail.isEnabled = false
        binding.tvResendTimer.visibility = View.VISIBLE

        resendTimer?.cancel()
        resendTimer = object : CountDownTimer(60000, 1000) {
            override fun onTick(millisUntilFinished: Long) {
                val seconds = millisUntilFinished / 1000
                binding.tvResendTimer.text = "Resend available in ${seconds}s"
            }

            override fun onFinish() {
                canResend = true
                binding.btnResendEmail.isEnabled = true
                binding.tvResendTimer.visibility = View.GONE
            }
        }.start()
    }

    private fun openEmailApp() {
        try {
            val intent = Intent(Intent.ACTION_MAIN)
            intent.addCategory(Intent.CATEGORY_APP_EMAIL)
            intent.addFlags(Intent.FLAG_ACTIVITY_NEW_TASK)
            startActivity(intent)
        } catch (e: Exception) {
            // Fallback: open Gmail or any email client
            try {
                val intent = Intent(Intent.ACTION_VIEW, Uri.parse("mailto:"))
                startActivity(intent)
            } catch (e: Exception) {
                showToast("No email app found")
            }
        }
    }

    private fun navigateToLogin() {
        startActivity(Intent(this, LoginActivity::class.java).apply {
            flags = Intent.FLAG_ACTIVITY_NEW_TASK or Intent.FLAG_ACTIVITY_CLEAR_TASK
        })
        finish()
    }

    private fun showLoading(isLoading: Boolean) {
        binding.apply {
            progressBar.visibility = if (isLoading) View.VISIBLE else View.GONE
            btnResendEmail.isEnabled = !isLoading && canResend
            btnOpenEmailApp.isEnabled = !isLoading
        }
    }

    override fun onDestroy() {
        super.onDestroy()
        resendTimer?.cancel()
    }
}