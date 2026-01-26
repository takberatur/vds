package com.agcforge.videodownloader.ui.activities

import android.annotation.SuppressLint
import android.content.Intent
import android.os.Bundle
import androidx.appcompat.app.AppCompatActivity
import androidx.core.splashscreen.SplashScreen.Companion.installSplashScreen
import androidx.lifecycle.lifecycleScope
import com.agcforge.videodownloader.R
import com.agcforge.videodownloader.ui.activities.auth.LoginActivity
import com.agcforge.videodownloader.utils.PreferenceManager
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.launch

@SuppressLint("CustomSplashScreen")
class SplashActivity : AppCompatActivity() {

    private lateinit var preferenceManager: PreferenceManager

    override fun onCreate(savedInstanceState: Bundle?) {
        // Install splash screen
        val splashScreen = installSplashScreen()

        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_splash)

        preferenceManager = PreferenceManager(this)

        // Keep the splash screen visible for this Activity
        splashScreen.setKeepOnScreenCondition { true }

        lifecycleScope.launch {
            // Simulate initialization
            delay(2000)

            // Check if user is logged in
            checkAuthenticationState()
        }
    }

    private suspend fun checkAuthenticationState() {
        val token = preferenceManager.authToken.first()

        if (!token.isNullOrEmpty()) {
            // User is logged in, go to main
            navigateToMain()
        } else {
            // User is not logged in, go to login
            navigateToLogin()
        }
    }

    private fun navigateToMain() {
        startActivity(Intent(this, MainActivity::class.java))
        finish()
    }

    private fun navigateToLogin() {
        startActivity(Intent(this, LoginActivity::class.java))
        finish()
    }
}