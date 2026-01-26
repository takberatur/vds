package com.agcforge.videodownloader.ui.activities

import android.annotation.SuppressLint
import android.content.Intent
import android.os.Bundle
import androidx.appcompat.app.AppCompatActivity
import androidx.core.splashscreen.SplashScreen.Companion.installSplashScreen
import androidx.lifecycle.lifecycleScope
import com.agcforge.videodownloader.R
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch

@SuppressLint("CustomSplashScreen")
class SplashActivity : AppCompatActivity() {

    override fun onCreate(savedInstanceState: Bundle?) {
        // Install splash screen
        val splashScreen = installSplashScreen()

        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_splash)

        // Keep the splash screen visible for this Activity
        splashScreen.setKeepOnScreenCondition { true }

        lifecycleScope.launch {
            // Simulate initialization
            delay(2000)

            // Navigate to main activity
            startActivity(Intent(this@SplashActivity, MainActivity::class.java))
            finish()
        }
    }
}