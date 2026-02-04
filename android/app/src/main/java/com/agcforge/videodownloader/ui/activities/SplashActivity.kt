package com.agcforge.videodownloader.ui.activities

import android.animation.Animator
import android.animation.ObjectAnimator
import android.annotation.SuppressLint
import android.content.Intent
import android.os.Bundle
import android.view.View
import android.view.animation.AccelerateDecelerateInterpolator
import androidx.appcompat.app.AppCompatActivity
import androidx.core.splashscreen.SplashScreen.Companion.installSplashScreen
import androidx.lifecycle.lifecycleScope
import com.agcforge.videodownloader.data.api.ApiClient
import com.agcforge.videodownloader.data.api.VideoDownloaderRepository
import com.agcforge.videodownloader.databinding.ActivitySplashBinding
import com.agcforge.videodownloader.helper.AdsConfig
import com.agcforge.videodownloader.helper.AdsConfigManager
import com.agcforge.videodownloader.utils.PreferenceManager
import com.airbnb.lottie.LottieAnimationView
import com.onesignal.OneSignal
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.launch

@SuppressLint("CustomSplashScreen")
class SplashActivity : AppCompatActivity() {

    private lateinit var binding: ActivitySplashBinding
    private lateinit var preferenceManager: PreferenceManager
    private val repository = VideoDownloaderRepository()

    override fun onCreate(savedInstanceState: Bundle?) {
        // Install splash screen
        installSplashScreen()

        super.onCreate(savedInstanceState)
        binding = ActivitySplashBinding.inflate(layoutInflater)
        setContentView(binding.root)

        preferenceManager = PreferenceManager(this)

        setupAnimations()
        startSplashSequence()
    }

    private fun setupAnimations() {
        binding.lottieAnimation.addAnimatorListener(object : Animator.AnimatorListener {
            override fun onAnimationStart(animation: Animator) {
                // Animation started
            }

            override fun onAnimationEnd(animation: Animator) {
                // Lottie animation completed
                animateTextElements()
            }

            override fun onAnimationCancel(animation: Animator) {}
            override fun onAnimationRepeat(animation: Animator) {}
        })
    }
    private fun startSplashSequence() {
        lifecycleScope.launch {
            // Wait for Lottie animation to complete (or set duration)
            delay(2500)

            lifecycleScope.launch {
                // Simulate initialization
                delay(2000)

                fetchAndStoreApplication()

                initializeAuthToken()

                AdsConfig.initialize(this@SplashActivity)
                val enableOneSignal: Boolean = AdsConfig.ONESIGNAL_ID != null
                if (enableOneSignal) OneSignal.initWithContext(this@SplashActivity, AdsConfig.ONESIGNAL_ID!!)
                if (enableOneSignal) {
                    OneSignal.Notifications.requestPermission(true)
                }

                navigateToMain()
            }
        }
    }

	private suspend fun fetchAndStoreApplication() {
		repository.getApplication()
			.onSuccess { app ->
				preferenceManager.saveApplication(app)
			}
			.onFailure {
				// ignore
			}
	}

	private suspend fun initializeAuthToken() {
		val token = preferenceManager.authToken.first()
		if (!token.isNullOrEmpty()) {
			ApiClient.setAuthToken(token)
		}
	}

    private fun navigateToMain() {
        binding.root.animate()
            .alpha(0f)
            .setDuration(300)
            .withEndAction {
                startActivity(Intent(this, MainActivity::class.java))
                overridePendingTransition(android.R.anim.fade_in, android.R.anim.fade_out)
                finish()
            }
            .start()
    }

    private fun animateTextElements() {
        // Animate app name
        ObjectAnimator.ofFloat(binding.tvAppName, View.ALPHA, 0f, 1f).apply {
            duration = 600
            startDelay = 200
            interpolator = AccelerateDecelerateInterpolator()
            start()
        }

        // Scale animation for app name
        binding.tvAppName.animate()
            .scaleX(1.1f)
            .scaleY(1.1f)
            .setDuration(300)
            .withEndAction {
                binding.tvAppName.animate()
                    .scaleX(1f)
                    .scaleY(1f)
                    .setDuration(300)
                    .start()
            }
            .start()
    }
}
