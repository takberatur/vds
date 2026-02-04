package com.agcforge.videodownloader

import android.annotation.SuppressLint
import android.app.Application
import com.agcforge.videodownloader.helper.AdsConfig
import com.agcforge.videodownloader.utils.DownloadManagerCleaner
import com.onesignal.OneSignal
import com.onesignal.debug.LogLevel
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch

class App : Application() {

    companion object {
        private const val TAG = "App_Root_Main"

        @SuppressLint("StaticFieldLeak")
        @Volatile
        private var mInstance: App? = null

        fun getInstance(): App {
            return mInstance ?: throw IllegalStateException("App not initialized")
        }
    }
    override fun onCreate() {
        super.onCreate()
        mInstance = this
		DownloadManagerCleaner.clearFailedDownloads(this)

        AdsConfig.initialize(this)

        OneSignal.Debug.logLevel = if (BuildConfig.DEBUG) LogLevel.VERBOSE else LogLevel.NONE
        AdsConfig.ONESIGNAL_ID?.let { OneSignal.initWithContext(this, it) }

        CoroutineScope(Dispatchers.IO).launch {
            OneSignal.Notifications.requestPermission(true)
        }
    }
}
